import React, { useState, useEffect } from 'react';
import { api } from '../api/client';
import { DashboardStats } from '../types';
import {
  FileText,
  CheckCircle2,
  Clock,
  AlertTriangle,
  UploadCloud,
  Search,
  Sparkles,
  ArrowRight,
  Loader2,
  FileCode,
  FileCheck,
} from 'lucide-react';

interface DashboardProps {
  onOpenUpload: () => void;
  onNavigate: (tab: string, docID?: string) => void;
}

export const Dashboard: React.FC<DashboardProps> = ({ onOpenUpload, onNavigate }) => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const resp = await api.dashboard.getStats();
      if (resp.success && resp.data) {
        setStats(resp.data);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to load dashboard metrics');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'Just now';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
    } catch {
      return dateStr;
    }
  };

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '60vh', gap: '1rem' }}>
        <Loader2 size={36} className="spinner" color="var(--brand-light)" />
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Loading operational metrics...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '2rem' }}>
        <div
          style={{
            padding: '1.5rem',
            backgroundColor: 'var(--danger-bg)',
            border: '1px solid var(--danger-border)',
            borderRadius: 'var(--radius-lg)',
            color: 'var(--danger)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <h3 style={{ fontSize: '1.1rem', marginBottom: '0.25rem' }}>Failed to load metrics</h3>
            <p style={{ fontSize: '0.85rem' }}>{error}</p>
          </div>
          <button onClick={fetchStats} className="btn-secondary">
            Retry
          </button>
        </div>
      </div>
    );
  }

  const awaitingProcessing = (stats?.queued_documents || 0) + (stats?.processing_documents || 0);

  return (
    <div className="fade-in" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* Welcome Banner */}
      <div
        className="glass-panel"
        style={{
          padding: '2.25rem',
          background: 'linear-gradient(180deg, rgba(79, 70, 229, 0.06) 0%, var(--bg-surface) 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          border: '1px solid rgba(79, 70, 229, 0.18)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div style={{ maxWidth: '620px', position: 'relative', zIndex: 2 }}>
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.4rem',
              fontSize: '0.75rem',
              fontWeight: 700,
              padding: '0.25rem 0.65rem',
              borderRadius: '999px',
              backgroundColor: 'rgba(79, 70, 229, 0.12)',
              color: 'var(--brand-light)',
              marginBottom: '0.75rem',
              border: '1px solid rgba(79, 70, 229, 0.25)',
            }}
          >
            <Sparkles size={14} />
            <span>AI Document Intelligence Active</span>
          </div>
          <h2 style={{ fontSize: '1.75rem', letterSpacing: '-0.02em', marginBottom: '0.5rem', color: 'var(--text-main)' }}>
            Operations & Document Intelligence
          </h2>
          <p style={{ fontSize: '0.92rem', color: 'var(--text-secondary)', lineHeight: 1.6 }}>
            Upload contracts, financial reports, and administrative files. AfriOps extracts text, creates vector search indexes, and powers grounded AI answers.
          </p>

          <div style={{ display: 'flex', gap: '0.85rem', marginTop: '1.25rem' }}>
            <button onClick={onOpenUpload} className="btn-primary">
              <UploadCloud size={16} />
              <span>Upload New Document</span>
            </button>
            <button onClick={() => onNavigate('search')} className="btn-secondary">
              <Search size={16} />
              <span>Search Passages</span>
            </button>
            <button onClick={() => onNavigate('chat')} className="btn-secondary">
              <Sparkles size={16} color="var(--brand-light)" />
              <span>Ask AI Question</span>
            </button>
          </div>
        </div>

        <div
          style={{
            position: 'absolute',
            right: '-40px',
            bottom: '-40px',
            width: '280px',
            height: '280px',
            background: 'radial-gradient(circle, rgba(79, 70, 229, 0.12) 0%, transparent 70%)',
            pointerEvents: 'none',
          }}
        />
      </div>

      {/* Metrics Row (PRD 4.E) */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.25rem' }}>
        {/* Total Documents */}
        <div className="glass-panel" style={{ padding: '1.5rem', position: 'relative' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <span style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
              Total Documents
            </span>
            <div
              style={{
                width: '36px',
                height: '36px',
                borderRadius: '8px',
                backgroundColor: 'rgba(99, 102, 241, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--brand-light)',
              }}
            >
              <FileText size={18} />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--text-main)' }}>
            {stats?.total_documents ?? 0}
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Active across organisation
          </p>
        </div>

        {/* Successfully Processed */}
        <div className="glass-panel" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <span style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
              Processed & Indexed
            </span>
            <div
              style={{
                width: '36px',
                height: '36px',
                borderRadius: '8px',
                backgroundColor: 'var(--success-bg)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--success)',
              }}
            >
              <CheckCircle2 size={18} />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--success)' }}>
            {stats?.completed_documents ?? 0}
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Ready for Q&A and search
          </p>
        </div>

        {/* Awaiting Processing */}
        <div className="glass-panel" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <span style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
              Awaiting Processing
            </span>
            <div
              style={{
                width: '36px',
                height: '36px',
                borderRadius: '8px',
                backgroundColor: 'var(--warning-bg)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--warning)',
              }}
            >
              <Clock size={18} />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--warning)' }}>
            {awaitingProcessing}
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Queued: {stats?.queued_documents ?? 0} • Running: {stats?.processing_documents ?? 0}
          </p>
        </div>

        {/* Failed Documents */}
        <div className="glass-panel" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <span style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
              Failed Extractions
            </span>
            <div
              style={{
                width: '36px',
                height: '36px',
                borderRadius: '8px',
                backgroundColor: 'var(--danger-bg)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--danger)',
              }}
            >
              <AlertTriangle size={18} />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--danger)' }}>
            {stats?.failed_documents ?? 0}
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
            Can be inspected and retried
          </p>
        </div>
      </div>

      {/* Recent Uploads Section */}
      <div className="glass-panel" style={{ padding: '1.75rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div>
            <h3 style={{ fontSize: '1.15rem', color: 'var(--text-main)', marginBottom: '0.2rem' }}>
              Recently Uploaded Documents
            </h3>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)' }}>
              Latest files queued and parsed by the processing pipeline
            </p>
          </div>
          <button
            onClick={() => onNavigate('documents')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              fontSize: '0.84rem',
              color: 'var(--brand-light)',
              fontWeight: 600,
            }}
          >
            <span>View All Documents</span>
            <ArrowRight size={14} />
          </button>
        </div>

        {!stats?.recent_documents || stats.recent_documents.length === 0 ? (
          <div
            style={{
              padding: '3.5rem 1.5rem',
              textAlign: 'center',
              backgroundColor: 'var(--bg-glass-subtle)',
              borderRadius: 'var(--radius-md)',
              border: '1px dashed var(--border-card)',
            }}
          >
            <FileCode size={40} color="var(--text-muted)" style={{ margin: '0 auto 1rem' }} />
            <h4 style={{ fontSize: '1rem', color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              No documents uploaded yet
            </h4>
            <p style={{ fontSize: '0.84rem', color: 'var(--text-muted)', marginBottom: '1.25rem' }}>
              Upload your first business document to start indexing and asking questions.
            </p>
            <button onClick={onOpenUpload} className="btn-primary">
              <UploadCloud size={16} />
              <span>Upload First Document</span>
            </button>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
            {stats.recent_documents.map((doc) => (
              <div
                key={doc.id}
                onClick={() => onNavigate('documents', doc.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '0.85rem 1.15rem',
                  backgroundColor: 'var(--bg-surface)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-md)',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = 'var(--bg-surface-hover)';
                  e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.15)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = 'var(--bg-surface)';
                  e.currentTarget.style.borderColor = 'var(--border-subtle)';
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.85rem' }}>
                  <div
                    style={{
                      width: '36px',
                      height: '36px',
                      borderRadius: '8px',
                      backgroundColor: 'rgba(99, 102, 241, 0.1)',
                      color: 'var(--brand-light)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <FileCheck size={18} />
                  </div>
                  <div>
                    <div style={{ fontSize: '0.92rem', fontWeight: 600, color: 'var(--text-main)' }}>
                      {doc.title}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'flex', gap: '0.6rem' }}>
                      <span>{doc.file_name}</span>
                      <span>•</span>
                      <span>{formatFileSize(doc.file_size)}</span>
                      <span>•</span>
                      <span>{formatDate(doc.created_at)}</span>
                    </div>
                  </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                  <span className={`status-pill status-${doc.status}`}>
                    {doc.status}
                  </span>
                  <ArrowRight size={14} color="var(--text-muted)" />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
