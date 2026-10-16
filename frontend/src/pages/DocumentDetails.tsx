import React, { useState, useEffect } from 'react';
import { api } from '../api/client';
import { Document, ProcessingStatus } from '../types';
import {
  ArrowLeft,
  Download,
  Edit,
  Trash2,
  Archive,
  ArchiveRestore,
  RefreshCw,
  FileText,
  Clock,
  Layers,
  FileCode,
  AlertTriangle,
  CheckCircle2,
  Tag,
  Hash,
  Loader2,
} from 'lucide-react';

interface DocumentDetailsProps {
  document: Document;
  onBack: () => void;
  onEdit: (doc: Document) => void;
  onRefreshList: () => void;
}

export const DocumentDetails: React.FC<DocumentDetailsProps> = ({
  document: initialDoc,
  onBack,
  onEdit,
  onRefreshList,
}) => {
  const [doc, setDoc] = useState<Document>(initialDoc);
  const [procStatus, setProcStatus] = useState<ProcessingStatus | null>(null);
  const [isLoadingStatus, setIsLoadingStatus] = useState(true);
  const [isRetrying, setIsRetrying] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const fetchDetails = async () => {
    try {
      const [docResp, statusResp] = await Promise.all([
        api.documents.getByID(doc.id),
        api.documents.getProcessingStatus(doc.id),
      ]);
      if (docResp.success && docResp.data) {
        setDoc(docResp.data);
      }
      if (statusResp.success && statusResp.data) {
        setProcStatus(statusResp.data);
      }
    } catch (err: any) {
      console.error('Failed to fetch details:', err);
    } finally {
      setIsLoadingStatus(false);
    }
  };

  useEffect(() => {
    fetchDetails();
    // Poll every 3 seconds if status is queued or processing
    if (doc.status === 'queued' || doc.status === 'processing') {
      const interval = setInterval(fetchDetails, 3000);
      return () => clearInterval(interval);
    }
  }, [doc.status]);

  const handleRetry = async () => {
    setIsRetrying(true);
    setActionError(null);
    try {
      const resp = await api.documents.retry(doc.id);
      if (resp.success && resp.data) {
        setDoc(resp.data);
        fetchDetails();
      }
    } catch (err: any) {
      setActionError(err?.message || 'Retry failed');
    } finally {
      setIsRetrying(false);
    }
  };

  const handleArchive = async () => {
    try {
      if (doc.is_archived) {
        const resp = await api.documents.unarchive(doc.id);
        if (resp.success) setDoc(resp.data);
      } else {
        const resp = await api.documents.archive(doc.id);
        if (resp.success) setDoc(resp.data);
      }
      onRefreshList();
    } catch (err: any) {
      alert(err?.message || 'Archive failed');
    }
  };

  const handleDelete = async () => {
    if (!window.confirm(`Are you sure you want to delete "${doc.title}"?`)) return;
    try {
      await api.documents.delete(doc.id);
      onRefreshList();
      onBack();
    } catch (err: any) {
      alert(err?.message || 'Delete failed');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  return (
    <div className="fade-in" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* Top Navigation */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <button
          onClick={onBack}
          className="btn-secondary"
          style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '0.55rem 1rem' }}
        >
          <ArrowLeft size={16} />
          <span>Back to Documents</span>
        </button>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <button onClick={() => onEdit(doc)} className="btn-secondary">
            <Edit size={16} />
            <span>Edit Metadata</span>
          </button>
          <a
            href={api.documents.downloadURL(doc.id)}
            download={doc.file_name}
            className="btn-secondary"
            style={{ display: 'inline-flex' }}
          >
            <Download size={16} />
            <span>Download</span>
          </a>
          <button onClick={handleArchive} className="btn-secondary">
            {doc.is_archived ? <ArchiveRestore size={16} /> : <Archive size={16} />}
            <span>{doc.is_archived ? 'Restore' : 'Archive'}</span>
          </button>
          <button
            onClick={handleDelete}
            className="btn-secondary"
            style={{ color: 'var(--danger)' }}
          >
            <Trash2 size={16} />
            <span>Delete</span>
          </button>
        </div>
      </div>

      {actionError && (
        <div
          style={{
            padding: '1rem',
            backgroundColor: 'var(--danger-bg)',
            border: '1px solid var(--danger-border)',
            borderRadius: 'var(--radius-md)',
            color: 'var(--danger)',
            fontSize: '0.86rem',
          }}
        >
          {actionError}
        </div>
      )}

      {/* Document Overview Header Card */}
      <div
        className="glass-panel"
        style={{
          padding: '2rem',
          display: 'flex',
          flexWrap: 'wrap',
          gap: '1.5rem',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <div
            style={{
              width: '56px',
              height: '56px',
              borderRadius: '14px',
              background: 'rgba(79, 70, 229, 0.12)',
              color: 'var(--brand-light)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '1px solid rgba(79, 70, 229, 0.25)',
            }}
          >
            <FileText size={28} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.25rem' }}>
              <h2 style={{ fontSize: '1.5rem', color: 'var(--text-main)', letterSpacing: '-0.02em' }}>
                {doc.title}
              </h2>
              <span className={`status-pill status-${doc.status}`}>{doc.status}</span>
              {doc.is_archived && (
                <span style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', borderRadius: '4px', background: 'rgba(255, 255, 255, 0.1)', color: 'var(--text-muted)' }}>
                  Archived
                </span>
              )}
            </div>
            <p style={{ fontSize: '0.86rem', color: 'var(--text-muted)' }}>
              Original File: <code style={{ color: 'var(--text-secondary)' }}>{doc.file_name}</code>
            </p>
          </div>
        </div>

        {/* Retry Button if Failed */}
        {doc.status === 'failed' && (
          <button
            onClick={handleRetry}
            className="btn-primary"
            disabled={isRetrying}
            style={{ background: '#d97706', borderColor: 'rgba(255, 255, 255, 0.15)' }}
          >
            {isRetrying ? (
              <>
                <Loader2 size={16} className="spinner" />
                <span>Re-triggering...</span>
              </>
            ) : (
              <>
                <RefreshCw size={16} />
                <span>Retry Processing</span>
              </>
            )}
          </button>
        )}
      </div>

      {/* Grid: Processing Pipeline Status & Document Attributes */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.5rem' }}>
        {/* Processing Pipeline Card */}
        <div className="glass-panel" style={{ padding: '1.75rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem' }}>
            <h3 style={{ fontSize: '1.1rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Layers size={18} color="var(--brand-light)" />
              <span>Processing & Extraction</span>
            </h3>
            <button onClick={fetchDetails} className="btn-ghost" title="Refresh">
              <RefreshCw size={14} />
            </button>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Status</span>
              <span className={`status-pill status-${doc.status}`}>{doc.status}</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Pipeline Step</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)', fontWeight: 600 }}>
                {procStatus?.job_type || 'text_extraction & chunking'}
              </span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Processing Attempts</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)' }}>
                {procStatus?.attempts || 1} / {procStatus?.max_attempts || 3}
              </span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Extracted Words</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)', fontWeight: 600 }}>
                {procStatus?.word_count ? procStatus.word_count.toLocaleString() : '—'}
              </span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Estimated Pages</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)' }}>
                {procStatus?.page_count || '1'}
              </span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Searchable Chunks</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--brand-light)', fontWeight: 600 }}>
                {procStatus?.chunk_count || '0'} chunks indexed
              </span>
            </div>

            {doc.error_message && (
              <div
                style={{
                  marginTop: '0.5rem',
                  padding: '0.85rem',
                  backgroundColor: 'var(--danger-bg)',
                  border: '1px solid var(--danger-border)',
                  borderRadius: 'var(--radius-md)',
                  color: 'var(--danger)',
                  fontSize: '0.8rem',
                  lineHeight: 1.5,
                }}
              >
                <strong>Error: </strong>
                {doc.error_message}
              </div>
            )}
          </div>
        </div>

        {/* File Attributes & Metadata Card */}
        <div className="glass-panel" style={{ padding: '1.75rem' }}>
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1.25rem' }}>
            <FileCode size={18} color="var(--brand-light)" />
            <span>File Attributes & Metadata</span>
          </h3>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>File Size</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)' }}>{formatFileSize(doc.file_size)}</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>MIME Type</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)' }}>{doc.mime_type}</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Uploaded Date</span>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-main)' }}>{new Date(doc.created_at).toLocaleString()}</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '0.75rem', borderBottom: '1px solid var(--border-subtle)' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>SHA-256 Hash</span>
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', fontFamily: 'monospace', maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {doc.file_hash}
              </span>
            </div>

            {/* Custom Metadata Tags */}
            <div style={{ marginTop: '0.5rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>Custom Tags</span>
                <button onClick={() => onEdit(doc)} style={{ fontSize: '0.75rem', color: 'var(--brand-light)', fontWeight: 600 }}>
                  + Edit Tags
                </button>
              </div>

              {!doc.metadata || Object.keys(doc.metadata).length === 0 ? (
                <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
                  No custom operational tags set.
                </p>
              ) : (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                  {Object.entries(doc.metadata).map(([k, v]) => (
                    <div
                      key={k}
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.35rem',
                        padding: '0.3rem 0.65rem',
                        background: 'rgba(99, 102, 241, 0.1)',
                        border: '1px solid rgba(99, 102, 241, 0.25)',
                        borderRadius: '6px',
                        fontSize: '0.78rem',
                      }}
                    >
                      <Tag size={12} color="var(--brand-light)" />
                      <strong style={{ color: 'var(--text-main)' }}>{k}:</strong>
                      <span style={{ color: 'var(--text-secondary)' }}>{v}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
