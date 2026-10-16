import React, { useState, useEffect, useRef } from 'react';
import { api } from '../api/client';
import { Conversation, Message } from '../types';
import {
  Sparkles,
  Send,
  Plus,
  MessageSquare,
  FileText,
  ChevronDown,
  ChevronUp,
  Loader2,
  Bot,
  User,
  Quote,
  ShieldCheck,
  BookOpen,
} from 'lucide-react';

interface AIChatProps {
  initialQuestion?: string;
}

export const AIChat: React.FC<AIChatProps> = ({ initialQuestion }) => {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConvID, setActiveConvID] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputQuestion, setInputQuestion] = useState(initialQuestion || '');
  const [isLoadingConvs, setIsLoadingConvs] = useState(true);
  const [isSending, setIsSending] = useState(false);
  const [expandedSources, setExpandedSources] = useState<Record<string, boolean>>({});
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 1. Fetch conversations
  const fetchConversations = async () => {
    try {
      const resp = await api.ai.listConversations();
      if (resp.success && resp.data) {
        setConversations(resp.data);
        if (resp.data.length > 0 && !activeConvID) {
          setActiveConvID(resp.data[0].id);
        }
      }
    } catch (err) {
      console.error('Failed to list conversations:', err);
    } finally {
      setIsLoadingConvs(false);
    }
  };

  useEffect(() => {
    fetchConversations();
  }, []);

  // 2. Fetch messages for active conversation
  const fetchMessages = async (convID: string) => {
    try {
      const resp = await api.ai.getMessages(convID);
      if (resp.success && resp.data) {
        setMessages(resp.data);
      }
    } catch (err) {
      console.error('Failed to get messages:', err);
    }
  };

  useEffect(() => {
    if (activeConvID) {
      fetchMessages(activeConvID);
    } else {
      setMessages([]);
    }
  }, [activeConvID]);

  // Auto-scroll to bottom of chat
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isSending]);

  // Handle new conversation
  const handleNewConversation = async () => {
    try {
      const resp = await api.ai.createConversation('New Chat');
      if (resp.success && resp.data) {
        setConversations([resp.data, ...conversations]);
        setActiveConvID(resp.data.id);
        setMessages([]);
      }
    } catch (err) {
      console.error('Failed to create conversation:', err);
    }
  };

  // Handle asking a question
  const handleSendMessage = async (e?: React.FormEvent, customQuestion?: string) => {
    if (e) e.preventDefault();
    const q = customQuestion || inputQuestion.trim();
    if (!q || isSending) return;

    let targetConvID = activeConvID;

    // Create a new conversation if none exists
    if (!targetConvID) {
      try {
        const title = q.length > 30 ? q.slice(0, 30) + '...' : q;
        const newConv = await api.ai.createConversation(title);
        if (newConv.success && newConv.data) {
          targetConvID = newConv.data.id;
          setConversations([newConv.data, ...conversations]);
          setActiveConvID(targetConvID);
        }
      } catch (err) {
        alert('Failed to initiate conversation');
        return;
      }
    }

    if (!targetConvID) return;

    // Optimistic user message
    const tempUserMsg: Message = {
      id: 'temp-' + Date.now(),
      conversation_id: targetConvID,
      sender_type: 'user',
      content: q,
      token_count: q.length / 4,
      created_at: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, tempUserMsg]);
    setInputQuestion('');
    setIsSending(true);

    try {
      const resp = await api.ai.ask(targetConvID, q);
      if (resp.success && resp.data) {
        setMessages((prev) => [...prev, resp.data]);
        // Update conversation in list if title was generic
        fetchConversations();
      }
    } catch (err: any) {
      const errorAssistantMsg: Message = {
        id: 'err-' + Date.now(),
        conversation_id: targetConvID,
        sender_type: 'assistant',
        content: `Sorry, an error occurred while generating the answer: ${err?.message || 'Please try again.'}`,
        token_count: 0,
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errorAssistantMsg]);
    } finally {
      setIsSending(false);
    }
  };

  const toggleSourceExpand = (msgID: string) => {
    setExpandedSources((prev) => ({ ...prev, [msgID]: !prev[msgID] }));
  };

  const suggestedPrompts = [
    'What are our main financial highlights from the latest reports?',
    'Summarize the key compliance and operational terms.',
    'Are there any pending contract deadlines or expiration dates?',
  ];

  return (
    <div
      style={{
        display: 'flex',
        height: 'calc(100vh - 70px)',
        overflow: 'hidden',
        backgroundColor: 'var(--bg-base)',
      }}
    >
      {/* Conversations Sidebar */}
      <div
        style={{
          width: '280px',
          borderRight: '1px solid var(--border-card)',
          backgroundColor: 'var(--bg-surface)',
          display: 'flex',
          flexDirection: 'column',
          flexShrink: 0,
        }}
      >
        <div style={{ padding: '1rem', borderBottom: '1px solid var(--border-subtle)' }}>
          <button
            onClick={handleNewConversation}
            className="btn-primary"
            style={{ width: '100%', padding: '0.65rem', justifyContent: 'center' }}
          >
            <Plus size={16} />
            <span>New Conversation</span>
          </button>
        </div>

        <div style={{ flex: 1, overflowY: 'auto', padding: '0.75rem' }}>
          <div style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', padding: '0.35rem 0.5rem' }}>
            Chat History
          </div>
          {conversations.length === 0 ? (
            <div style={{ padding: '2rem 1rem', textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.82rem' }}>
              No conversations yet. Start one above!
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem', marginTop: '0.4rem' }}>
              {conversations.map((conv) => {
                const isActive = conv.id === activeConvID;
                return (
                  <button
                    key={conv.id}
                    onClick={() => setActiveConvID(conv.id)}
                    style={{
                      width: '100%',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.65rem',
                      padding: '0.65rem 0.75rem',
                      borderRadius: 'var(--radius-md)',
                      backgroundColor: isActive ? 'rgba(99, 102, 241, 0.15)' : 'transparent',
                      color: isActive ? 'var(--brand-light)' : 'var(--text-secondary)',
                      border: isActive ? '1px solid rgba(99, 102, 241, 0.3)' : '1px solid transparent',
                      textAlign: 'left',
                      fontSize: '0.86rem',
                      fontWeight: isActive ? 600 : 500,
                    }}
                  >
                    <MessageSquare size={15} style={{ flexShrink: 0 }} />
                    <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {conv.title}
                    </span>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Main Chat Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
        {/* Messages Stream */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '2rem 3rem', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {messages.length === 0 ? (
            <div style={{ margin: 'auto', maxWidth: '580px', textAlign: 'center', padding: '2rem 0' }}>
              <div
                style={{
                  width: '56px',
                  height: '56px',
                  borderRadius: '16px',
                  background: 'var(--brand-gradient)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                  margin: '0 auto 1.25rem',
                  boxShadow: 'var(--brand-glow)',
                }}
              >
                <Sparkles size={28} />
              </div>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--text-main)', marginBottom: '0.5rem' }}>
                AfriOps Document Assistant
              </h2>
              <p style={{ fontSize: '0.88rem', color: 'var(--text-secondary)', lineHeight: 1.6, marginBottom: '2rem' }}>
                Ask any question grounded strictly in your organisation's uploaded business documents. Answers include source citations identifying exact files and passages.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
                <p style={{ fontSize: '0.78rem', color: 'var(--text-muted)', fontWeight: 600, textTransform: 'uppercase' }}>
                  Suggested Questions
                </p>
                {suggestedPrompts.map((p, i) => (
                  <button
                    key={i}
                    onClick={() => handleSendMessage(undefined, p)}
                    style={{
                      padding: '0.75rem 1rem',
                      backgroundColor: 'var(--bg-surface)',
                      border: '1px solid var(--border-card)',
                      borderRadius: 'var(--radius-md)',
                      color: 'var(--text-main)',
                      fontSize: '0.86rem',
                      textAlign: 'left',
                      transition: 'all 0.15s ease',
                      cursor: 'pointer',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = 'var(--bg-surface-hover)';
                      e.currentTarget.style.borderColor = 'var(--border-focus)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'var(--bg-surface)';
                      e.currentTarget.style.borderColor = 'var(--border-card)';
                    }}
                  >
                    "{p}"
                  </button>
                ))}
              </div>
            </div>
          ) : (
            messages.map((msg) => {
              const isUser = msg.sender_type === 'user';
              const isExpanded = !!expandedSources[msg.id];
              return (
                <div
                  key={msg.id}
                  style={{
                    display: 'flex',
                    gap: '1rem',
                    alignSelf: isUser ? 'flex-end' : 'flex-start',
                    maxWidth: '85%',
                  }}
                >
                  {!isUser && (
                    <div
                      style={{
                        width: '36px',
                        height: '36px',
                        borderRadius: '10px',
                        background: 'var(--brand-gradient)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'white',
                        flexShrink: 0,
                      }}
                    >
                      <Bot size={20} />
                    </div>
                  )}

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <div
                      style={{
                        padding: '1rem 1.25rem',
                        borderRadius: 'var(--radius-lg)',
                        backgroundColor: isUser ? 'var(--brand-dark)' : 'var(--bg-surface)',
                        color: 'var(--text-main)',
                        border: isUser ? 'none' : '1px solid var(--border-card)',
                        fontSize: '0.92rem',
                        lineHeight: 1.6,
                        whiteSpace: 'pre-wrap',
                      }}
                    >
                      {msg.content}
                    </div>

                    {/* Grounded Source Citations (PRD 4.D) */}
                    {!isUser && msg.sources && msg.sources.length > 0 && (
                      <div
                        style={{
                          backgroundColor: 'var(--bg-glass-subtle)',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 'var(--radius-md)',
                          padding: '0.65rem 0.85rem',
                        }}
                      >
                        <button
                          onClick={() => toggleSourceExpand(msg.id)}
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            width: '100%',
                            color: 'var(--brand-light)',
                            fontSize: '0.78rem',
                            fontWeight: 600,
                          }}
                        >
                          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                            <ShieldCheck size={14} color="var(--success)" />
                            <span>Grounded in {msg.sources.length} document citations</span>
                          </div>
                          {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                        </button>

                        {isExpanded && (
                          <div style={{ marginTop: '0.65rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                            {msg.sources.map((src, i) => (
                              <div
                                key={src.id || i}
                                style={{
                                  padding: '0.65rem',
                                  backgroundColor: 'var(--bg-base)',
                                  borderRadius: '6px',
                                  border: '1px solid var(--border-subtle)',
                                  fontSize: '0.78rem',
                                }}
                              >
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                                  <BookOpen size={12} color="var(--brand-light)" />
                                  <span>{src.document_title || src.file_name || 'Document'}</span>
                                  {src.page_number && <span style={{ color: 'var(--text-muted)' }}>• Page {src.page_number}</span>}
                                </div>
                                <p style={{ color: 'var(--text-secondary)', fontStyle: 'italic', lineHeight: 1.4 }}>
                                  "{src.citation_text}"
                                </p>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>

                  {isUser && (
                    <div
                      style={{
                        width: '36px',
                        height: '36px',
                        borderRadius: '10px',
                        backgroundColor: 'var(--bg-elevated)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'var(--text-secondary)',
                        flexShrink: 0,
                      }}
                    >
                      <User size={18} />
                    </div>
                  )}
                </div>
              );
            })
          )}

          {isSending && (
            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', color: 'var(--text-muted)', fontSize: '0.86rem' }}>
              <div
                style={{
                  width: '36px',
                  height: '36px',
                  borderRadius: '10px',
                  background: 'var(--brand-gradient)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                }}
              >
                <Bot size={20} />
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Loader2 size={16} className="spinner" color="var(--brand-light)" />
                <span>Reading relevant passages & synthesizing answer...</span>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Bar */}
        <div style={{ padding: '1.25rem 3rem', borderTop: '1px solid var(--border-card)', backgroundColor: 'var(--bg-surface)' }}>
          <form onSubmit={handleSendMessage} style={{ display: 'flex', gap: '0.75rem', position: 'relative' }}>
            <input
              type="text"
              placeholder="Ask a question about your documents..."
              value={inputQuestion}
              onChange={(e) => setInputQuestion(e.target.value)}
              disabled={isSending}
              style={{
                flex: 1,
                padding: '0.85rem 1.25rem',
                fontSize: '0.92rem',
                backgroundColor: 'var(--bg-base)',
                borderRadius: 'var(--radius-lg)',
              }}
            />
            <button
              type="submit"
              className="btn-primary"
              disabled={!inputQuestion.trim() || isSending}
              style={{ padding: '0.85rem 1.5rem', borderRadius: 'var(--radius-lg)' }}
            >
              {isSending ? <Loader2 size={18} className="spinner" /> : <Send size={18} />}
            </button>
          </form>
          <p style={{ fontSize: '0.72rem', color: 'var(--text-muted)', textAlign: 'center', marginTop: '0.5rem' }}>
            Answers are strictly grounded in your active organisation's documents. Other organisations cannot see your data.
          </p>
        </div>
      </div>
    </div>
  );
};
