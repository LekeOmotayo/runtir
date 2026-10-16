import React from 'react';
import { UploadCloud, Search, CheckCircle2 } from 'lucide-react';

interface HeaderProps {
  title: string;
  subtitle?: string;
  onOpenUpload: () => void;
  onSearchClick?: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  title,
  subtitle,
  onOpenUpload,
  onSearchClick,
}) => {
  return (
    <header
      style={{
        height: '70px',
        borderBottom: '1px solid var(--border-card)',
        backgroundColor: 'var(--bg-glass)',
        backdropFilter: 'blur(16px)',
        WebkitBackdropFilter: 'blur(16px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 2rem',
        position: 'sticky',
        top: 0,
        zIndex: 40,
      }}
    >
      {/* Title & Context */}
      <div>
        <h1
          style={{
            fontSize: '1.25rem',
            fontWeight: 700,
            color: 'var(--text-main)',
            letterSpacing: '-0.02em',
          }}
        >
          {title}
        </h1>
        {subtitle && (
          <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            {subtitle}
          </p>
        )}
      </div>

      {/* Right Controls */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        {/* Quick Search Trigger */}
        {onSearchClick && (
          <button
            onClick={onSearchClick}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.65rem',
              backgroundColor: 'var(--bg-surface)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-md)',
              padding: '0.45rem 0.85rem',
              color: 'var(--text-secondary)',
              fontSize: '0.82rem',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = 'var(--border-focus)';
              e.currentTarget.style.color = 'var(--text-main)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = 'var(--border-subtle)';
              e.currentTarget.style.color = 'var(--text-secondary)';
            }}
          >
            <Search size={14} color="var(--brand-light)" />
            <span>Search document content...</span>
            <kbd
              style={{
                background: 'rgba(255, 255, 255, 0.08)',
                padding: '0.1rem 0.35rem',
                borderRadius: '4px',
                fontSize: '0.7rem',
                color: 'var(--text-muted)',
              }}
            >
              /
            </kbd>
          </button>
        )}

        {/* System Health Dot */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
            fontSize: '0.75rem',
            color: 'var(--text-muted)',
            padding: '0.35rem 0.65rem',
            background: 'var(--bg-glass-subtle)',
            borderRadius: '999px',
            border: '1px solid var(--border-subtle)',
          }}
          title="Backend API connected and healthy"
        >
          <span
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              backgroundColor: 'var(--success)',
              boxShadow: '0 0 8px var(--success)',
              display: 'inline-block',
            }}
          />
          <span>API v1 Connected</span>
        </div>

        {/* Upload Action */}
        <button
          onClick={onOpenUpload}
          className="btn-primary"
          style={{ padding: '0.55rem 1.15rem' }}
        >
          <UploadCloud size={16} />
          <span>Upload Document</span>
        </button>
      </div>
    </header>
  );
};
