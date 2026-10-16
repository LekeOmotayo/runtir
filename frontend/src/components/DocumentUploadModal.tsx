import React, { useState, useRef } from 'react';
import { X, UploadCloud, FileText, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';
import { api } from '../api/client';
import { Document } from '../types';

interface DocumentUploadModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (newDoc: Document) => void;
}

export const DocumentUploadModal: React.FC<DocumentUploadModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState('');
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const allowedExtensions = ['.pdf', '.txt', '.docx', '.doc', '.md'];

  const validateAndSetFile = (selectedFile: File) => {
    setError(null);
    const ext = '.' + selectedFile.name.split('.').pop()?.toLowerCase();
    if (!allowedExtensions.includes(ext)) {
      setError(`Unsupported file format. Please upload PDF, TXT, DOCX, DOC, or MD.`);
      return;
    }
    if (selectedFile.size > 25 * 1024 * 1024) {
      setError('File size exceeds the 25 MB limit.');
      return;
    }
    setFile(selectedFile);
    if (!title) {
      setTitle(selectedFile.name.replace(/\.[^/.]+$/, ''));
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      validateAndSetFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      validateAndSetFile(e.target.files[0]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setError('Please choose a file to upload.');
      return;
    }

    setIsUploading(true);
    setError(null);

    try {
      const resp = await api.documents.upload(file, title.trim());
      if (resp.success && resp.data) {
        setUploadSuccess(true);
        setTimeout(() => {
          onSuccess(resp.data);
          handleClose();
        }, 1000);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to upload document. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  const handleClose = () => {
    setFile(null);
    setTitle('');
    setError(null);
    setUploadSuccess(false);
    onClose();
  };

  return (
    <div className="modal-overlay" onClick={handleClose}>
      <div
        className="glass-panel fade-in"
        style={{
          width: '100%',
          maxWidth: '520px',
          padding: '2rem',
          backgroundColor: 'var(--bg-surface)',
          border: '1px solid var(--border-card)',
          position: 'relative',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <div>
            <h2 style={{ fontSize: '1.25rem', color: 'var(--text-main)' }}>Upload Business Document</h2>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)' }}>
              Files are automatically parsed, indexed, and made searchable by AI.
            </p>
          </div>
          <button onClick={handleClose} className="btn-ghost" style={{ padding: '0.35rem' }}>
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

        {uploadSuccess ? (
          <div style={{ textAlign: 'center', padding: '2.5rem 0' }}>
            <CheckCircle2 size={48} color="var(--success)" style={{ margin: '0 auto 1rem' }} />
            <h3 style={{ fontSize: '1.15rem', color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Document Queued for Processing!
            </h3>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
              Text extraction and AI vector indexing have been triggered in the background.
            </p>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            {/* Drag and Drop Box */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              style={{
                border: isDragging
                  ? '2px dashed var(--brand-light)'
                  : '2px dashed var(--border-card)',
                backgroundColor: isDragging ? 'rgba(99, 102, 241, 0.08)' : 'var(--bg-glass-subtle)',
                borderRadius: 'var(--radius-lg)',
                padding: '2.5rem 1.5rem',
                textAlign: 'center',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                marginBottom: '1.5rem',
              }}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.txt,.docx,.doc,.md"
                onChange={handleFileChange}
                style={{ display: 'none' }}
              />

              {file ? (
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.75rem' }}>
                  <div
                    style={{
                      width: '42px',
                      height: '42px',
                      borderRadius: '10px',
                      backgroundColor: 'rgba(99, 102, 241, 0.15)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: 'var(--brand-light)',
                    }}
                  >
                    <FileText size={22} />
                  </div>
                  <div style={{ textAlign: 'left' }}>
                    <div style={{ fontSize: '0.92rem', fontWeight: 600, color: 'var(--text-main)' }}>
                      {file.name}
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      {(file.size / (1024 * 1024)).toFixed(2)} MB • Click to replace
                    </div>
                  </div>
                </div>
              ) : (
                <>
                  <div
                    style={{
                      width: '48px',
                      height: '48px',
                      borderRadius: '50%',
                      background: 'rgba(99, 102, 241, 0.12)',
                      color: 'var(--brand-light)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      margin: '0 auto 1rem',
                    }}
                  >
                    <UploadCloud size={24} />
                  </div>
                  <p style={{ fontSize: '0.92rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                    Click to select or drag & drop document
                  </p>
                  <p style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                    Supported formats: PDF, TXT, DOCX, DOC, MD (Max 25 MB)
                  </p>
                </>
              )}
            </div>

            {/* Custom Title Input */}
            <div style={{ marginBottom: '1.75rem' }}>
              <label
                style={{
                  display: 'block',
                  fontSize: '0.82rem',
                  fontWeight: 600,
                  color: 'var(--text-secondary)',
                  marginBottom: '0.4rem',
                }}
              >
                Document Title (Optional)
              </label>
              <input
                type="text"
                placeholder="e.g. Q3 2026 Financial Report"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                style={{ width: '100%' }}
                disabled={isUploading}
              />
            </div>

            {/* Submit & Cancel Buttons */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button
                type="button"
                onClick={handleClose}
                className="btn-secondary"
                disabled={isUploading}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn-primary"
                disabled={!file || isUploading}
                style={{ minWidth: '130px' }}
              >
                {isUploading ? (
                  <>
                    <Loader2 size={16} className="spinner" />
                    <span>Uploading...</span>
                  </>
                ) : (
                  <>
                    <UploadCloud size={16} />
                    <span>Upload</span>
                  </>
                )}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};
