import React, { useState, useEffect } from 'react';
import { X, Plus, Trash2, Save, Loader2, AlertCircle } from 'lucide-react';
import { Document } from '../types';
import { api } from '../api/client';

interface DocumentEditModalProps {
  document: Document | null;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (updatedDoc: Document) => void;
}

export const DocumentEditModal: React.FC<DocumentEditModalProps> = ({
  document,
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [title, setTitle] = useState('');
  const [metadataEntries, setMetadataEntries] = useState<Array<{ key: string; value: string }>>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (document) {
      setTitle(document.title);
      if (document.metadata) {
        setMetadataEntries(
          Object.entries(document.metadata).map(([key, value]) => ({ key, value }))
        );
      } else {
        setMetadataEntries([]);
      }
    }
  }, [document]);

  if (!isOpen || !document) return null;

  const handleAddMeta = () => {
    setMetadataEntries([...metadataEntries, { key: '', value: '' }]);
  };

  const handleRemoveMeta = (index: number) => {
    setMetadataEntries(metadataEntries.filter((_, i) => i !== index));
  };

  const handleMetaChange = (index: number, field: 'key' | 'value', val: string) => {
    const updated = [...metadataEntries];
    updated[index][field] = val;
    setMetadataEntries(updated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    const metaMap: Record<string, string> = {};
    for (const entry of metadataEntries) {
      if (entry.key.trim()) {
        metaMap[entry.key.trim()] = entry.value.trim();
      }
    }

    try {
      const resp = await api.documents.update(document.id, {
        title: title.trim(),
        metadata: metaMap,
      });
      if (resp.success && resp.data) {
        onSuccess(resp.data);
        onClose();
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to update document metadata');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className="glass-panel fade-in"
        style={{
          width: '100%',
          maxWidth: '560px',
          padding: '2rem',
          backgroundColor: 'var(--bg-surface)',
          border: '1px solid var(--border-card)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <div>
            <h2 style={{ fontSize: '1.25rem', color: 'var(--text-main)' }}>Edit Document & Metadata</h2>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              Update display title and custom operational metadata tags.
            </p>
          </div>
          <button onClick={onClose} className="btn-ghost" style={{ padding: '0.35rem' }}>
            <X size={20} />
          </button>
        </div>

        {error && (
          <div
            style={{
              padding: '0.75rem 1rem',
              backgroundColor: 'var(--danger-bg)',
              border: '1px solid var(--danger-border)',
              borderRadius: 'var(--radius-md)',
              color: 'var(--danger)',
              fontSize: '0.84rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              marginBottom: '1.25rem',
            }}
          >
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {/* Title input */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{ display: 'block', fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>
              Document Title
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
              style={{ width: '100%' }}
            />
          </div>

          {/* Custom Metadata Key-Values */}
          <div style={{ marginBottom: '1.75rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.6rem' }}>
              <label style={{ fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
                Metadata Tags (Key-Value)
              </label>
              <button
                type="button"
                onClick={handleAddMeta}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.35rem',
                  fontSize: '0.78rem',
                  color: 'var(--brand-light)',
                  fontWeight: 600,
                }}
              >
                <Plus size={14} />
                <span>Add Tag</span>
              </button>
            </div>

            {metadataEntries.length === 0 ? (
              <div
                style={{
                  padding: '1.25rem',
                  backgroundColor: 'var(--bg-glass-subtle)',
                  borderRadius: 'var(--radius-md)',
                  textAlign: 'center',
                  color: 'var(--text-muted)',
                  fontSize: '0.82rem',
                  border: '1px dashed var(--border-card)',
                }}
              >
                No custom metadata tags added. Click "Add Tag" above to specify department, category, fiscal year, etc.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.65rem' }}>
                {metadataEntries.map((entry, idx) => (
                  <div key={idx} style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <input
                      type="text"
                      placeholder="Key (e.g. Department)"
                      value={entry.key}
                      onChange={(e) => handleMetaChange(idx, 'key', e.target.value)}
                      style={{ flex: 1, fontSize: '0.85rem' }}
                    />
                    <input
                      type="text"
                      placeholder="Value (e.g. Finance)"
                      value={entry.value}
                      onChange={(e) => handleMetaChange(idx, 'value', e.target.value)}
                      style={{ flex: 1.5, fontSize: '0.85rem' }}
                    />
                    <button
                      type="button"
                      onClick={() => handleRemoveMeta(idx)}
                      style={{
                        padding: '0.5rem',
                        color: 'var(--text-muted)',
                        borderRadius: 'var(--radius-sm)',
                      }}
                      onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--danger)')}
                      onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Action buttons */}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
            <button type="button" onClick={onClose} className="btn-secondary" disabled={isSubmitting}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 size={16} className="spinner" />
                  <span>Saving...</span>
                </>
              ) : (
                <>
                  <Save size={16} />
                  <span>Save Changes</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
