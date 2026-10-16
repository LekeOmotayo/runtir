import React, { useState, useEffect } from 'react';
import { api } from '../api/client';
import { Document } from '../types';
import {
  Search,
  Filter,
  Download,
  Trash2,
  Archive,
  ArchiveRestore,
  Edit,
  Eye,
  Loader2,
  FileText,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  MoreVertical,
  Check,
  RefreshCw,
} from 'lucide-react';

interface DocumentsProps {
  onOpenUpload: () => void;
  onSelectDocument: (doc: Document) => void;
  onEditDocument: (doc: Document) => void;
}

export const Documents: React.FC<DocumentsProps> = ({
  onOpenUpload,
  onSelectDocument,
  onEditDocument,
}) => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(15);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [mimeFilter, setMimeFilter] = useState('');
  const [isArchived, setIsArchived] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionMenuOpen, setActionMenuOpen] = useState<string | null>(null);

  const fetchDocuments = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const resp = await api.documents.list({
        q: searchQuery.trim() || undefined,
        status: statusFilter || undefined,
        mime_type: mimeFilter || undefined,
        is_archived: isArchived,
        page,
        page_size: pageSize,
      });
      if (resp.success) {
        setDocuments(resp.data || []);
        setTotal(resp.pagination?.total || 0);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to retrieve documents');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDocuments();
  }, [page, statusFilter, mimeFilter, isArchived]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      setPage(1);
      fetchDocuments();
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const handleArchive = async (doc: Document) => {
    try {
      if (doc.is_archived) {
        await api.documents.unarchive(doc.id);
      } else {
        await api.documents.archive(doc.id);
      }
      fetchDocuments();
    } catch (err: any) {
      alert(err?.message || 'Archive operation failed');
    }
  };

  const handleDelete = async (doc: Document) => {
    if (!window.confirm(`Are you sure you want to delete "${doc.title}"? This cannot be undone.`)) {
      return;
    }
    try {
      await api.documents.delete(doc.id);
      fetchDocuments();
    } catch (err: any) {
      alert(err?.message || 'Delete operation failed');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const totalPages = Math.ceil(total / pageSize) || 1;

  return (
    <div className="fade-in" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Top Filter Bar */}
      <div
        className="glass-panel"
        style={{
          padding: '1.25rem 1.5rem',
          display: 'flex',
          flexWrap: 'wrap',
          gap: '1rem',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        {/* Search Input */}
        <div style={{ position: 'relative', flex: '1 1 280px', maxWidth: '400px' }}>
          <Search
            size={16}
            color="var(--text-muted)"
            style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)' }}
          />
          <input
            type="text"
            placeholder="Search by title or filename..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: '100%', paddingLeft: '2.4rem' }}
          />
        </div>

        {/* Filter Controls */}
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem', alignItems: 'center' }}>
          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              setPage(1);
            }}
            style={{ minWidth: '140px' }}
          >
            <option value="">All Statuses</option>
            <option value="completed">Completed</option>
            <option value="processing">Processing</option>
            <option value="queued">Queued</option>
            <option value="failed">Failed</option>
          </select>

          {/* Format Filter */}
          <select
            value={mimeFilter}
            onChange={(e) => {
              setMimeFilter(e.target.value);
              setPage(1);
            }}
            style={{ minWidth: '140px' }}
          >
            <option value="">All Formats</option>
            <option value="application/pdf">PDF Document</option>
            <option value="application/vnd.openxmlformats-officedocument.wordprocessingml.document">Word (.docx)</option>
            <option value="text/plain">Plain Text (.txt)</option>
            <option value="text/markdown">Markdown (.md)</option>
          </select>

          {/* Archive Toggle */}
          <button
            onClick={() => {
              setIsArchived(!isArchived);
              setPage(1);
            }}
            className={isArchived ? 'btn-primary' : 'btn-secondary'}
            style={{ fontSize: '0.82rem', padding: '0.65rem 0.95rem' }}
          >
            <Archive size={14} />
            <span>{isArchived ? 'Archived Docs' : 'Active Docs'}</span>
          </button>

          {/* Refresh Button */}
          <button onClick={fetchDocuments} className="btn-ghost" title="Refresh">
            <RefreshCw size={16} />
          </button>
        </div>
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

      {/* Documents Table */}
      <div className="glass-panel" style={{ overflow: 'hidden' }}>
        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '5rem 0', gap: '0.75rem' }}>
            <Loader2 size={32} className="spinner" color="var(--brand-light)" />
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>Loading documents...</p>
          </div>
        ) : documents.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '5rem 2rem' }}>
            <FileText size={48} color="var(--text-muted)" style={{ margin: '0 auto 1rem' }} />
            <h3 style={{ fontSize: '1.15rem', color: 'var(--text-main)', marginBottom: '0.35rem' }}>
              No documents found
            </h3>
            <p style={{ fontSize: '0.86rem', color: 'var(--text-muted)', marginBottom: '1.5rem', maxWidth: '400px', margin: '0 auto 1.5rem' }}>
              {searchQuery || statusFilter || mimeFilter
                ? 'No documents matched the specified filters. Try clearing filters.'
                : 'Upload your company documents to start processing and indexing.'}
            </p>
            <button onClick={onOpenUpload} className="btn-primary">
              Upload Document
            </button>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.88rem' }}>
              <thead>
                <tr
                  style={{
                    borderBottom: '1px solid var(--border-card)',
                    backgroundColor: 'rgba(255, 255, 255, 0.02)',
                    color: 'var(--text-muted)',
                    fontSize: '0.78rem',
                    textTransform: 'uppercase',
                    letterSpacing: '0.04em',
                  }}
                >
                  <th style={{ padding: '0.9rem 1.5rem' }}>Document</th>
                  <th style={{ padding: '0.9rem 1rem' }}>Status</th>
                  <th style={{ padding: '0.9rem 1rem' }}>Format</th>
                  <th style={{ padding: '0.9rem 1rem' }}>Size</th>
                  <th style={{ padding: '0.9rem 1rem' }}>Uploaded</th>
                  <th style={{ padding: '0.9rem 1.5rem', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {documents.map((doc) => (
                  <tr
                    key={doc.id}
                    style={{
                      borderBottom: '1px solid var(--border-subtle)',
                      transition: 'background 0.15s ease',
                    }}
                    onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = 'var(--bg-glass-subtle)')}
                    onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                  >
                    {/* Title & Filename */}
                    <td style={{ padding: '1rem 1.5rem' }}>
                      <div
                        onClick={() => onSelectDocument(doc)}
                        style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.75rem' }}
                      >
                        <div
                          style={{
                            width: '34px',
                            height: '34px',
                            borderRadius: '8px',
                            backgroundColor: 'rgba(99, 102, 241, 0.12)',
                            color: 'var(--brand-light)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            flexShrink: 0,
                          }}
                        >
                          <FileText size={18} />
                        </div>
                        <div>
                          <div style={{ fontWeight: 600, color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                            <span>{doc.title}</span>
                            {doc.is_archived && (
                              <span style={{ fontSize: '0.68rem', padding: '0.1rem 0.4rem', borderRadius: '4px', backgroundColor: 'rgba(255, 255, 255, 0.1)', color: 'var(--text-muted)' }}>
                                Archived
                              </span>
                            )}
                          </div>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{doc.file_name}</div>
                        </div>
                      </div>
                    </td>

                    {/* Status */}
                    <td style={{ padding: '1rem' }}>
                      <span className={`status-pill status-${doc.status}`}>
                        {doc.status}
                      </span>
                    </td>

                    {/* Format */}
                    <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
                      {doc.file_name.split('.').pop()?.toUpperCase() || 'FILE'}
                    </td>

                    {/* Size */}
                    <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
                      {formatFileSize(doc.file_size)}
                    </td>

                    {/* Date */}
                    <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
                      {new Date(doc.created_at).toLocaleDateString()}
                    </td>

                    {/* Actions */}
                    <td style={{ padding: '1rem 1.5rem', textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
                        <button
                          onClick={() => onSelectDocument(doc)}
                          className="btn-ghost"
                          title="View Details"
                          style={{ padding: '0.45rem' }}
                        >
                          <Eye size={16} />
                        </button>
                        <button
                          onClick={() => onEditDocument(doc)}
                          className="btn-ghost"
                          title="Edit Title & Metadata"
                          style={{ padding: '0.45rem' }}
                        >
                          <Edit size={16} />
                        </button>
                        <a
                          href={api.documents.downloadURL(doc.id)}
                          download={doc.file_name}
                          className="btn-ghost"
                          title="Download"
                          style={{ padding: '0.45rem', display: 'inline-flex' }}
                        >
                          <Download size={16} />
                        </a>
                        <button
                          onClick={() => handleArchive(doc)}
                          className="btn-ghost"
                          title={doc.is_archived ? 'Restore Document' : 'Archive Document'}
                          style={{ padding: '0.45rem' }}
                        >
                          {doc.is_archived ? <ArchiveRestore size={16} /> : <Archive size={16} />}
                        </button>
                        <button
                          onClick={() => handleDelete(doc)}
                          className="btn-ghost"
                          title="Delete"
                          style={{ padding: '0.45rem' }}
                          onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--danger)')}
                          onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Bar */}
        {total > pageSize && (
          <div
            style={{
              padding: '1rem 1.5rem',
              borderTop: '1px solid var(--border-card)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              fontSize: '0.84rem',
              color: 'var(--text-secondary)',
            }}
          >
            <div>
              Showing {((page - 1) * pageSize) + 1} to {Math.min(page * pageSize, total)} of {total} documents
            </div>
            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.65rem' }}
              >
                <ChevronLeft size={16} />
              </button>
              <span style={{ fontWeight: 600 }}>
                {page} / {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.65rem' }}
              >
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
