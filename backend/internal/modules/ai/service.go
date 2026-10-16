package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
)

const systemPrompt = `You are AfriOps Assistant, an AI that answers questions strictly based on the provided document context.

Rules:
1. Only answer using information from the CONTEXT section below.
2. If the context does not contain enough information to answer the question, say: "The available documents do not contain enough information to answer this question."
3. Always cite the source document name and chunk when answering.
4. Be concise, accurate, and professional.
5. Never speculate or add information not present in the context.
`

type ConversationDTO struct {
	ID    uuid.UUID    `json:"id"`
	OrgID uuid.UUID    `json:"organisation_id"`
	Title string       `json:"title"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type MessageDTO struct {
	ID             uuid.UUID    `json:"id"`
	ConversationID uuid.UUID    `json:"conversation_id"`
	SenderType     string       `json:"sender_type"`
	Content        string       `json:"content"`
	TokenCount     int32        `json:"token_count"`
	Sources        []SourceDTO  `json:"sources,omitempty"`
	CreatedAt      string       `json:"created_at"`
}

type SourceDTO struct {
	ID             uuid.UUID `json:"id"`
	DocumentID     uuid.UUID `json:"document_id,omitempty"`
	ChunkID        uuid.UUID `json:"chunk_id,omitempty"`
	DocumentTitle  string    `json:"document_title,omitempty"`
	FileName       string    `json:"file_name,omitempty"`
	PageNumber     int32     `json:"page_number,omitempty"`
	RelevanceScore float64   `json:"relevance_score"`
	CitationText   string    `json:"citation_text"`
}

type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

type CreateConversationRequest struct {
	Title string `json:"title"`
}

type Service struct {
	cfg        *config.Config
	queries    *db.Queries
	embedder   EmbeddingProvider
	completer  CompletionProvider
}

func NewService(cfg *config.Config, queries *db.Queries) *Service {
	return &Service{
		cfg:       cfg,
		queries:   queries,
		embedder:  NewEmbeddingProvider(),
		completer: NewCompletionProvider(),
	}
}

// CreateConversation starts a new AI chat thread for a user in the organisation
func (s *Service) CreateConversation(ctx context.Context, orgID, userID uuid.UUID, title string) (*ConversationDTO, error) {
	if strings.TrimSpace(title) == "" {
		title = "New Conversation"
	}

	conv, err := s.queries.CreateConversation(ctx, db.CreateConversationParams{
		OrganisationID: orgID,
		UserID:         userID,
		Title:          title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return toConversationDTO(conv), nil
}

// ListConversations returns paginated conversations for a user
func (s *Service) ListConversations(ctx context.Context, orgID, userID uuid.UUID, page, pageSize int64) ([]ConversationDTO, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 || pageSize > 50 { pageSize = 20 }
	offset := (page - 1) * pageSize

	convs, err := s.queries.ListConversationsByUser(ctx, db.ListConversationsByUserParams{
		OrganisationID: orgID,
		UserID:         userID,
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	var dtos []ConversationDTO
	for _, c := range convs {
		dtos = append(dtos, *toConversationDTO(c))
	}
	return dtos, nil
}

// GetMessages returns all messages in a conversation (must belong to org)
func (s *Service) GetMessages(ctx context.Context, orgID, convID uuid.UUID) ([]MessageDTO, error) {
	conv, err := s.queries.GetConversationByID(ctx, db.GetConversationByIDParams{
		ID:             convID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}
	_ = conv

	messages, err := s.queries.ListMessagesByConversation(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	var dtos []MessageDTO
	for _, m := range messages {
		dto := MessageDTO{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			SenderType:     m.SenderType,
			Content:        m.Content,
			TokenCount:     m.TokenCount,
			CreatedAt:      formatTime(m.CreatedAt),
		}

		// Fetch sources for assistant messages
		if m.SenderType == "assistant" {
			sources, _ := s.queries.ListMessageSources(ctx, m.ID)
			for _, src := range sources {
				dto.Sources = append(dto.Sources, toSourceDTO(src))
			}
		}

		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// Ask runs the full RAG pipeline: embed question -> retrieve chunks -> generate answer -> save citations
func (s *Service) Ask(ctx context.Context, orgID, userID, convID uuid.UUID, question string) (*MessageDTO, error) {
	// 1. Verify conversation belongs to org
	_, err := s.queries.GetConversationByID(ctx, db.GetConversationByIDParams{
		ID:             convID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("conversation not found or access denied")
	}

	// 2. Save user message
	userMsg, err := s.queries.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: convID,
		SenderType:     "user",
		Content:        question,
		TokenCount:     int32(len(question) / 4),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}
	_ = userMsg

	// 3. Embed the question
	embeddings, err := s.embedder.Embed(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("failed to embed question: %w", err)
	}
	questionVec := embeddings[0]

	// 4. Retrieve all chunks for this organisation
	// (Production: use pgvector ANN search; MVP: in-memory cosine similarity)
	chunks, err := s.queries.SearchChunksByContent(ctx, db.SearchChunksByContentParams{
		OrganisationID: orgID,
		Column2:        pgtype.Text{String: question, Valid: true},
		Limit:          50,
		Offset:         0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	// 5. Score chunks using keyword overlap (placeholder when embeddings are dummies)
	type scoredChunk struct {
		chunk db.DocumentChunk
		score float64
	}
	var scored []scoredChunk

	for _, chunk := range chunks {
		var score float64
		if len(questionVec) > 0 && questionVec[0] != 0 {
			// Real embeddings available: embed chunk and compute cosine sim
			chunkEmbs, embErr := s.embedder.Embed(ctx, []string{chunk.Content})
			if embErr == nil && len(chunkEmbs) > 0 {
				score = cosineSimilarity(questionVec, chunkEmbs[0])
			}
		} else {
			// Placeholder mode: score by keyword overlap
			score = keywordScore(question, chunk.Content)
		}
		scored = append(scored, scoredChunk{chunk: chunk, score: score})
	}

	// Sort by score descending, take top 5
	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	if len(scored) > 5 {
		scored = scored[:5]
	}

	// 6. Build context for LLM
	var contextParts []string
	for i, sc := range scored {
		contextParts = append(contextParts, fmt.Sprintf("[Source %d, Doc: %s]\n%s", i+1, sc.chunk.DocumentID, sc.chunk.Content))
	}
	context := strings.Join(contextParts, "\n\n---\n\n")

	fullSystemPrompt := systemPrompt + "\n\nCONTEXT:\n" + context
	if len(contextParts) == 0 {
		fullSystemPrompt = systemPrompt + "\n\nCONTEXT:\n[No relevant documents found for this question]"
	}

	// 7. Generate AI answer
	answer, tokenCount, err := s.completer.Complete(ctx, fullSystemPrompt, question)
	if err != nil {
		log.Printf("[AI Chat Error] Completion failed: %v", err)
		answer = "An error occurred while generating the AI response. Please try again."
		tokenCount = 0
	}

	// 8. Save AI assistant message
	assistantMsg, err := s.queries.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: convID,
		SenderType:     "assistant",
		Content:        answer,
		TokenCount:     int32(tokenCount),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// 9. Save citations
	var sourceDTOs []SourceDTO
	for _, sc := range scored {
		docUUID := pgtype.UUID{Bytes: sc.chunk.DocumentID, Valid: true}
		chunkUUID := pgtype.UUID{Bytes: sc.chunk.ID, Valid: true}

		citationText := sc.chunk.Content
		if len(citationText) > 300 {
			citationText = citationText[:300] + "..."
		}

		src, dbErr := s.queries.CreateMessageSource(ctx, db.CreateMessageSourceParams{
			MessageID:      assistantMsg.ID,
			ChunkID:        chunkUUID,
			DocumentID:     docUUID,
			RelevanceScore: sc.score,
			CitationText:   citationText,
		})
		if dbErr == nil {
			sourceDTOs = append(sourceDTOs, SourceDTO{
				ID:             src.ID,
				DocumentID:     sc.chunk.DocumentID,
				ChunkID:        sc.chunk.ID,
				RelevanceScore: sc.score,
				CitationText:   citationText,
			})
		}
	}

	// 10. Update conversation timestamp
	_, _ = s.queries.TouchConversation(ctx, convID)

	return &MessageDTO{
		ID:             assistantMsg.ID,
		ConversationID: convID,
		SenderType:     "assistant",
		Content:        answer,
		TokenCount:     assistantMsg.TokenCount,
		Sources:        sourceDTOs,
		CreatedAt:      formatTime(assistantMsg.CreatedAt),
	}, nil
}

// GenerateEmbeddingsForDocument embeds all chunks of a document
func (s *Service) GenerateEmbeddingsForDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	chunks, err := s.queries.ListChunksByDocument(ctx, db.ListChunksByDocumentParams{
		DocumentID:     docID,
		OrganisationID: orgID,
	})
	if err != nil || len(chunks) == 0 {
		return nil
	}

	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	embeddings, err := s.embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	for i, chunk := range chunks {
		if i >= len(embeddings) {
			break
		}
		vecJSON, _ := json.Marshal(embeddings[i])
		_, _ = s.queries.CreateEmbedding(ctx, db.CreateEmbeddingParams{
			ChunkID:        chunk.ID,
			EmbeddingModel: getEnvOrPlaceholder("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
			Embedding:      vecJSON,
		})
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func keywordScore(question, content string) float64 {
	q := strings.ToLower(question)
	c := strings.ToLower(content)
	words := strings.Fields(q)
	if len(words) == 0 {
		return 0
	}
	hits := 0
	for _, w := range words {
		if len(w) > 3 && strings.Contains(c, w) {
			hits++
		}
	}
	return float64(hits) / float64(len(words))
}

func toConversationDTO(c db.AiConversation) *ConversationDTO {
	return &ConversationDTO{
		ID:        c.ID,
		OrgID:     c.OrganisationID,
		Title:     c.Title,
		CreatedAt: formatTime(c.CreatedAt),
		UpdatedAt: formatTime(c.UpdatedAt),
	}
}

func toSourceDTO(src db.ListMessageSourcesRow) SourceDTO {
	dto := SourceDTO{
		ID:             src.ID,
		RelevanceScore: src.RelevanceScore,
		CitationText:   src.CitationText,
	}
	if src.DocumentID.Valid {
		dto.DocumentID = src.DocumentID.Bytes
	}
	if src.ChunkID.Valid {
		dto.ChunkID = src.ChunkID.Bytes
	}
	if src.DocumentTitle.Valid {
		dto.DocumentTitle = src.DocumentTitle.String
	}
	if src.DocumentFileName.Valid {
		dto.FileName = src.DocumentFileName.String
	}
	if src.PageNumber.Valid {
		dto.PageNumber = src.PageNumber.Int32
	}
	return dto
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}
