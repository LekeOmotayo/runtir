export interface User {
  id: string;
  email: string;
  full_name: string;
  avatar_url?: string;
}

export interface Organisation {
  id: string;
  name: string;
  slug: string;
  role: 'owner' | 'admin' | 'member';
  created_at?: string;
  updated_at?: string;
}

export interface Member {
  id: string;
  user_id: string;
  email: string;
  full_name: string;
  avatar_url?: string;
  role: 'owner' | 'admin' | 'member';
  created_at: string;
}

export interface Document {
  id: string;
  organisation_id: string;
  uploaded_by?: string;
  title: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  file_hash: string;
  status: 'queued' | 'processing' | 'completed' | 'failed';
  error_message?: string;
  is_archived: boolean;
  archived_at?: string;
  metadata?: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface ContentSearchResult {
  chunk_id: string;
  document_id: string;
  document_title: string;
  file_name: string;
  mime_type: string;
  chunk_index: number;
  page_number?: number;
  snippet: string;
}

export interface ProcessingStatus {
  document_id: string;
  status: string;
  error_message?: string;
  job_type?: string;
  attempts: number;
  max_attempts: number;
  started_at?: string;
  completed_at?: string;
  page_count: number;
  word_count: number;
  chunk_count: number;
}

export interface Conversation {
  id: string;
  organisation_id: string;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface SourceCitation {
  id: string;
  document_id?: string;
  chunk_id?: string;
  document_title?: string;
  file_name?: string;
  page_number?: number;
  relevance_score: number;
  citation_text: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_type: 'user' | 'assistant';
  content: string;
  token_count: number;
  sources?: SourceCitation[];
  created_at: string;
}

export interface DashboardStats {
  total_documents: number;
  completed_documents: number;
  queued_documents: number;
  processing_documents: number;
  failed_documents: number;
  recent_documents: Array<{
    id: string;
    title: string;
    file_name: string;
    mime_type: string;
    file_size: number;
    status: string;
    created_at: string;
  }>;
}

export interface Pagination {
  page: number;
  page_size: number;
  total: number;
}

export interface PaginatedResponse<T> {
  success: boolean;
  message: string;
  data: T[];
  pagination: Pagination;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  error?: string;
}
