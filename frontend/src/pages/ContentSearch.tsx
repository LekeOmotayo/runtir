import React, { useState, useEffect } from 'react';
import { api } from '../api/client';
import { ContentSearchResult } from '../types';
import {
  Search,
  FileText,
  Sparkles,
  ArrowRight,
  Loader2,
  AlertCircle,
  FileSearch,
  BookOpen,
} from 'lucide-react';

interface ContentSearchProps {
  onAskAIAboutPassage: (passage: string, docTitle: string) => void;
  onNavigateToDoc: (docID: string) => void;
}

export const ContentSearch: React.FC<ContentSearchProps> = ({
  onAskAIAboutPassage,
  onNavigateToDoc,
}) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<ContentSearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(15);
  const [isLoading, setIsLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSearch = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const clean = query.trim();
    if (!clean) return;

    setIsLoading(true);
    setError(null);
    setHasSearched(true);

    try {
      const resp = await api.documents.searchContent(clean, page, pageSize);
      if (resp.success) {
        setResults(resp.data || []);
        setTotal(resp.pagination?.total || 0);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to search document contents');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (hasSearched && query.trim()) {
      handleSearch();
    }
  }, [page]);

  return (
    <div className="fade-in" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* Search Header Banner */}
      <div
        className="glass-panel"
        style={{
          padding: '2.5rem',
          textAlign: 'center',
          background: 'linear-gradient(180deg, rgba(79, 70, 229, 0.06) 0%, var(--bg-surface) 100%)',
          border: '1px solid rgba(79, 70, 229, 0.18)',
        }}
      >
        <div
          style={{
            width: '48px',
            height: '48px',
            borderRadius: '12px',
            background: 'var(--brand-primary)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            margin: '0 auto 1rem',
            boxShadow: '0 2px 8px rgba(79, 70, 229, 0.25)',
          }}
        >
          <Search size={24} />
        </div>
        <h2 style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--text-main)', marginBottom: '0.4rem' }}>
          Deep Document Content Search
        </h2>
        <p style={{ fontSize: '0.92rem', color: 'var(--text-secondary)', maxWidth: '580px', margin: '0 auto 1.75rem' }}>
          Search directly inside the extracted text of all uploaded contracts, invoices, and reports across your organisation.
        </p>

        {/* Search Bar Input */}
        <form onSubmit={handleSearch} style={{ maxWidth: '650px', margin: '0 auto', display: 'flex', gap: '0.75rem' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search
              size={18}
              color="var(--text-muted)"
              style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}
            />
            <input
              type="text"
              placeholder="e.g. indemnity clause, EBITDA margin, tax exemption, termination terms..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              style={{
                width: '100%',
                padding: '0.85rem 1rem 0.85rem 2.75rem',
                fontSize: '0.95rem',
                backgroundColor: 'var(--bg-surface)',
                border: '1px solid var(--border-card)',
              }}
            />
          </div>
          <button type="submit" className="btn-primary" disabled={isLoading || !query.trim()} style={{ padding: '0.85rem 1.5rem' }}>
            {isLoading ? <Loader2 size={18} className="spinner" /> : <span>Search</span>}
          </button>
        </form>
      </div>

      {error && (
        <div
          style={{
            padding: '1rem',
            backgroundColor: 'var(--danger-bg)',
            border: '1px solid var(--danger-border)',
            borderRadius: 'var(--radius-md)',
            color: 'var(--danger)',
            fontSize: '0.86rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <AlertCircle size={16} />
          <span>{error}</span>
        </div>
      )}

      {/* Results Section */}
      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '4rem 0', gap: '0.75rem' }}>
          <Loader2 size={32} className="spinner" color="var(--brand-light)" />
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Searching extracted passages...</p>
        </div>
      ) : hasSearched && results.length === 0 ? (
        <div className="glass-panel" style={{ padding: '4rem 2rem', textAlign: 'center' }}>
          <FileSearch size={44} color="var(--text-muted)" style={{ margin: '0 auto 1rem' }} />
          <h3 style={{ fontSize: '1.15rem', color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            No matching passages found
          </h3>
          <p style={{ fontSize: '0.86rem', color: 'var(--text-muted)' }}>
            Try adjusting your search terms or verify that documents have completed processing.
          </p>
        </div>
      ) : results.length > 0 ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
            <span>Found <strong>{total}</strong> relevant passages for "{query}"</span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
            {results.map((res) => (
              <div
                key={res.chunk_id}
                className="glass-panel"
                style={{
                  padding: '1.5rem',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.85rem',
                  transition: 'border-color 0.2s ease',
                }}
              >
                {/* Passage Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div
                    onClick={() => onNavigateToDoc(res.document_id)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.5rem',
                      cursor: 'pointer',
                      color: 'var(--brand-light)',
                      fontWeight: 600,
                      fontSize: '0.92rem',
                    }}
                  >
                    <BookOpen size={16} />
                    <span>{res.document_title}</span>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>({res.file_name})</span>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                    {res.page_number && <span>Page {res.page_number} • </span>}
                    <span>Passage #{res.chunk_index + 1}</span>
                  </div>
                </div>

                {/* Excerpt Snippet */}
                <div
                  style={{
                    padding: '1rem',
                    backgroundColor: 'var(--bg-base)',
                    borderRadius: 'var(--radius-md)',
                    border: '1px solid var(--border-subtle)',
                    fontSize: '0.88rem',
                    lineHeight: 1.6,
                    color: 'var(--text-secondary)',
                    fontFamily: 'monospace',
                  }}
                >
                  "{res.snippet}"
                </div>

                {/* Action Shortcuts */}
                <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                  <button
                    onClick={() => onAskAIAboutPassage(res.snippet, res.document_title)}
                    className="btn-secondary"
                    style={{ fontSize: '0.8rem', padding: '0.45rem 0.85rem' }}
                  >
                    <Sparkles size={14} color="var(--brand-light)" />
                    <span>Ask AI About This Passage</span>
                  </button>
                  <button
                    onClick={() => onNavigateToDoc(res.document_id)}
                    className="btn-ghost"
                    style={{ fontSize: '0.8rem', padding: '0.45rem 0.85rem' }}
                  >
                    <span>View Document</span>
                    <ArrowRight size={14} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
};
