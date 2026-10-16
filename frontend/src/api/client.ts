import {
  ApiResponse,
  PaginatedResponse,
  User,
  Organisation,
  Member,
  Document,
  ContentSearchResult,
  ProcessingStatus,
  Conversation,
  Message,
  DashboardStats,
} from '../types';

const BASE_URL = '/api/v1';

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers || {});

  // If body is not FormData, set Content-Type to application/json
  if (options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const token = localStorage.getItem('afriops_token');
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const config: RequestInit = {
    ...options,
    headers,
    credentials: 'include', // Includes HTTP-only cookies
  };

  const response = await fetch(url, config);

  if (response.status === 401 && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/register')) {
    // Optionally trigger session refresh or redirect
    localStorage.removeItem('afriops_token');
  }

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const errorMsg = data?.error || data?.message || `Request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  return data;
}

export const api = {
  // Auth
  auth: {
    register: (payload: { email: string; password: string; full_name: string; org_name?: string }) =>
      request<ApiResponse<{ token: string; user: User; active_organisation: Organisation; organisations: Organisation[] }>>(
        '/auth/register',
        { method: 'POST', body: JSON.stringify(payload) }
      ),
    login: (payload: { email: string; password: string; org_id?: string }) =>
      request<ApiResponse<{ token: string; user: User; active_organisation: Organisation; organisations: Organisation[] }>>(
        '/auth/login',
        { method: 'POST', body: JSON.stringify(payload) }
      ),
    logout: () =>
      request<ApiResponse<null>>('/auth/logout', { method: 'POST' }),
    me: () =>
      request<ApiResponse<{ user: User; active_organisation: Organisation; organisations: Organisation[] }>>(
        '/auth/me',
        { method: 'GET' }
      ),
    switchOrg: (org_id: string) =>
      request<ApiResponse<{ token: string; user: User; active_organisation: Organisation; organisations: Organisation[] }>>(
        '/auth/switch-org',
        { method: 'POST', body: JSON.stringify({ org_id }) }
      ),
    refresh: () =>
      request<ApiResponse<{ token: string; user: User; active_organisation: Organisation; organisations: Organisation[] }>>(
        '/auth/refresh',
        { method: 'POST' }
      ),
  },

  // Documents
  documents: {
    list: (params: { q?: string; status?: string; mime_type?: string; is_archived?: boolean; page?: number; page_size?: number } = {}) => {
      const searchParams = new URLSearchParams();
      if (params.q) searchParams.set('q', params.q);
      if (params.status) searchParams.set('status', params.status);
      if (params.mime_type) searchParams.set('mime_type', params.mime_type);
      if (params.is_archived !== undefined) searchParams.set('is_archived', String(params.is_archived));
      if (params.page) searchParams.set('page', String(params.page));
      if (params.page_size) searchParams.set('page_size', String(params.page_size));
      const qs = searchParams.toString();
      return request<PaginatedResponse<Document>>(`/documents${qs ? `?${qs}` : ''}`);
    },

    searchContent: (query: string, page = 1, pageSize = 20) =>
      request<PaginatedResponse<ContentSearchResult>>(
        `/documents/search?q=${encodeURIComponent(query)}&page=${page}&page_size=${pageSize}`
      ),

    getByID: (id: string) =>
      request<ApiResponse<Document>>(`/documents/${id}`),

    upload: (file: File, title?: string) => {
      const formData = new FormData();
      formData.append('file', file);
      if (title) formData.append('title', title);
      return request<ApiResponse<Document>>('/documents/upload', {
        method: 'POST',
        body: formData,
      });
    },

    update: (id: string, payload: { title?: string; metadata?: Record<string, string> }) =>
      request<ApiResponse<Document>>(`/documents/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(payload),
      }),

    archive: (id: string) =>
      request<ApiResponse<Document>>(`/documents/${id}/archive`, { method: 'POST' }),

    unarchive: (id: string) =>
      request<ApiResponse<Document>>(`/documents/${id}/unarchive`, { method: 'POST' }),

    getProcessingStatus: (id: string) =>
      request<ApiResponse<ProcessingStatus>>(`/documents/${id}/processing`),

    retry: (id: string) =>
      request<ApiResponse<Document>>(`/documents/${id}/retry`, { method: 'POST' }),

    delete: (id: string) =>
      request<ApiResponse<null>>(`/documents/${id}`, { method: 'DELETE' }),

    downloadURL: (id: string) => `${BASE_URL}/documents/${id}/download`,
  },

  // AI Assistant (RAG)
  ai: {
    listConversations: (page = 1, pageSize = 20) =>
      request<ApiResponse<Conversation[]>>(`/ai/conversations?page=${page}&page_size=${pageSize}`),

    createConversation: (title?: string) =>
      request<ApiResponse<Conversation>>('/ai/conversations', {
        method: 'POST',
        body: JSON.stringify({ title: title || 'New Conversation' }),
      }),

    getMessages: (conversationID: string) =>
      request<ApiResponse<Message[]>>(`/ai/conversations/${conversationID}/messages`),

    ask: (conversationID: string, question: string) =>
      request<ApiResponse<Message>>(`/ai/conversations/${conversationID}/ask`, {
        method: 'POST',
        body: JSON.stringify({ question }),
      }),
  },

  // Dashboard stats
  dashboard: {
    getStats: () =>
      request<ApiResponse<DashboardStats>>('/dashboard/stats'),
  },

  // Organisations & Members
  organisations: {
    getCurrent: () =>
      request<ApiResponse<Organisation>>('/organisations/current'),

    listMembers: () =>
      request<ApiResponse<Member[]>>('/organisations/members'),

    addMember: (payload: { email: string; full_name?: string; role: 'owner' | 'admin' | 'member' }) =>
      request<ApiResponse<Member>>('/organisations/members', {
        method: 'POST',
        body: JSON.stringify(payload),
      }),

    updateMemberRole: (userID: string, role: 'owner' | 'admin' | 'member') =>
      request<ApiResponse<Member>>(`/organisations/members/${userID}`, {
        method: 'PATCH',
        body: JSON.stringify({ role }),
      }),

    removeMember: (userID: string) =>
      request<ApiResponse<null>>(`/organisations/members/${userID}`, {
        method: 'DELETE',
      }),
  },
};
