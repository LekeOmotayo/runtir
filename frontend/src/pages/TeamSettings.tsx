import React, { useState, useEffect } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Member, Organisation } from '../types';
import {
  Users,
  Building2,
  UserPlus,
  Shield,
  Trash2,
  Loader2,
  AlertCircle,
  CheckCircle2,
  X,
  Mail,
  User as UserIcon,
} from 'lucide-react';

export const TeamSettings: React.FC = () => {
  const { activeOrg, user } = useAuth();
  const [orgDetails, setOrgDetails] = useState<Organisation | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Add Member Modal State
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [newEmail, setNewEmail] = useState('');
  const [newFullName, setNewFullName] = useState('');
  const [newRole, setNewRole] = useState<'owner' | 'admin' | 'member'>('member');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [modalError, setModalError] = useState<string | null>(null);

  const fetchData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [orgResp, memResp] = await Promise.all([
        api.organisations.getCurrent(),
        api.organisations.listMembers(),
      ]);
      if (orgResp.success) setOrgDetails(orgResp.data);
      if (memResp.success) setMembers(memResp.data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load team data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setModalError(null);

    try {
      const resp = await api.organisations.addMember({
        email: newEmail.trim(),
        full_name: newFullName.trim() || undefined,
        role: newRole,
      });
      if (resp.success && resp.data) {
        setMembers([...members, resp.data]);
        setIsAddModalOpen(false);
        setNewEmail('');
        setNewFullName('');
        setNewRole('member');
      }
    } catch (err: any) {
      setModalError(err?.message || 'Failed to add member');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRoleChange = async (targetUserID: string, role: 'owner' | 'admin' | 'member') => {
    try {
      const resp = await api.organisations.updateMemberRole(targetUserID, role);
      if (resp.success && resp.data) {
        setMembers(members.map((m) => (m.user_id === targetUserID ? resp.data : m)));
      }
    } catch (err: any) {
      alert(err?.message || 'Failed to update member role');
    }
  };

  const handleRemoveMember = async (m: Member) => {
    if (!window.confirm(`Are you sure you want to remove ${m.full_name || m.email} from the organisation?`)) {
      return;
    }

    try {
      await api.organisations.removeMember(m.user_id);
      setMembers(members.filter((mem) => mem.user_id !== m.user_id));
    } catch (err: any) {
      alert(err?.message || 'Failed to remove member');
    }
  };

  const isPrivileged = activeOrg?.role === 'owner' || activeOrg?.role === 'admin';

  return (
    <div className="fade-in" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* Organisation Profile Header */}
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
              width: '54px',
              height: '54px',
              borderRadius: '14px',
              background: 'rgba(79, 70, 229, 0.12)',
              color: 'var(--brand-light)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '1px solid rgba(79, 70, 229, 0.25)',
            }}
          >
            <Building2 size={26} />
          </div>
          <div>
            <h2 style={{ fontSize: '1.5rem', color: 'var(--text-main)', letterSpacing: '-0.02em', marginBottom: '0.2rem' }}>
              {orgDetails?.name || activeOrg?.name}
            </h2>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.8rem', fontSize: '0.82rem', color: 'var(--text-muted)' }}>
              <span>Slug: <code style={{ color: 'var(--text-secondary)' }}>{orgDetails?.slug || activeOrg?.slug}</code></span>
              <span>•</span>
              <span>Your Role: <strong style={{ color: 'var(--brand-light)', textTransform: 'capitalize' }}>{activeOrg?.role}</strong></span>
              <span>•</span>
              <span>Members: {members.length}</span>
            </div>
          </div>
        </div>

        {isPrivileged && (
          <button onClick={() => setIsAddModalOpen(true)} className="btn-primary">
            <UserPlus size={16} />
            <span>Invite Team Member</span>
          </button>
        )}
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
          }}
        >
          {error}
        </div>
      )}

      {/* Members Table */}
      <div className="glass-panel" style={{ overflow: 'hidden' }}>
        <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-card)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h3 style={{ fontSize: '1.1rem', color: 'var(--text-main)', marginBottom: '0.2rem' }}>
              Team Members & Access Roles
            </h3>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              Manage users who can view, upload, search, and query documents for this organisation.
            </p>
          </div>
        </div>

        {isLoading ? (
          <div style={{ padding: '4rem 0', textAlign: 'center' }}>
            <Loader2 size={32} className="spinner" color="var(--brand-light)" style={{ margin: '0 auto 0.5rem' }} />
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.86rem' }}>Loading team members...</p>
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
                  <th style={{ padding: '0.85rem 1.5rem' }}>User</th>
                  <th style={{ padding: '0.85rem 1rem' }}>Role</th>
                  <th style={{ padding: '0.85rem 1rem' }}>Joined</th>
                  <th style={{ padding: '0.85rem 1.5rem', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {members.map((m) => {
                  const isCurrentUser = m.user_id === user?.id;
                  return (
                    <tr
                      key={m.id}
                      style={{
                        borderBottom: '1px solid var(--border-subtle)',
                        transition: 'background 0.15s ease',
                      }}
                    >
                      <td style={{ padding: '1rem 1.5rem' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                          <div
                            style={{
                              width: '34px',
                              height: '34px',
                              borderRadius: '50%',
                              backgroundColor: 'var(--bg-elevated)',
                              border: '1px solid var(--border-card)',
                              color: 'var(--brand-light)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontWeight: 700,
                              fontSize: '0.85rem',
                            }}
                          >
                            {m.full_name ? m.full_name[0].toUpperCase() : m.email[0].toUpperCase()}
                          </div>
                          <div>
                            <div style={{ fontWeight: 600, color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                              <span>{m.full_name || 'Member'}</span>
                              {isCurrentUser && (
                                <span style={{ fontSize: '0.68rem', padding: '0.1rem 0.35rem', borderRadius: '4px', background: 'rgba(99, 102, 241, 0.2)', color: 'var(--brand-light)' }}>
                                  You
                                </span>
                              )}
                            </div>
                            <div style={{ fontSize: '0.76rem', color: 'var(--text-muted)' }}>{m.email}</div>
                          </div>
                        </div>
                      </td>

                      <td style={{ padding: '1rem' }}>
                        {isPrivileged && !isCurrentUser ? (
                          <select
                            value={m.role}
                            onChange={(e) => handleRoleChange(m.user_id, e.target.value as any)}
                            style={{ padding: '0.4rem 0.75rem', fontSize: '0.82rem' }}
                          >
                            <option value="owner">Owner</option>
                            <option value="admin">Admin</option>
                            <option value="member">Member</option>
                          </select>
                        ) : (
                          <span
                            style={{
                              textTransform: 'capitalize',
                              padding: '0.25rem 0.65rem',
                              borderRadius: '999px',
                              fontSize: '0.76rem',
                              fontWeight: 600,
                              backgroundColor: m.role === 'owner' ? 'rgba(99, 102, 241, 0.15)' : 'var(--bg-glass-subtle)',
                              color: m.role === 'owner' ? 'var(--brand-light)' : 'var(--text-secondary)',
                              border: '1px solid var(--border-subtle)',
                            }}
                          >
                            {m.role}
                          </span>
                        )}
                      </td>

                      <td style={{ padding: '1rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
                        {m.created_at ? new Date(m.created_at).toLocaleDateString() : 'Active'}
                      </td>

                      <td style={{ padding: '1rem 1.5rem', textAlign: 'right' }}>
                        {isPrivileged && !isCurrentUser && (
                          <button
                            onClick={() => handleRemoveMember(m)}
                            className="btn-ghost"
                            title="Remove Member"
                            style={{ padding: '0.45rem' }}
                            onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--danger)')}
                            onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
                          >
                            <Trash2 size={16} />
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Invite Member Modal */}
      {isAddModalOpen && (
        <div className="modal-overlay" onClick={() => setIsAddModalOpen(false)}>
          <div
            className="glass-panel fade-in"
            style={{
              width: '100%',
              maxWidth: '480px',
              padding: '2rem',
              backgroundColor: 'var(--bg-surface)',
              border: '1px solid var(--border-card)',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <div>
                <h2 style={{ fontSize: '1.25rem', color: 'var(--text-main)' }}>Add Team Member</h2>
                <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                  Give access to this organisation's documents and AI knowledge.
                </p>
              </div>
              <button onClick={() => setIsAddModalOpen(false)} className="btn-ghost" style={{ padding: '0.35rem' }}>
                <X size={20} />
              </button>
            </div>

            {modalError && (
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
                <span>{modalError}</span>
              </div>
            )}

            <form onSubmit={handleAddMember}>
              <div style={{ marginBottom: '1.25rem' }}>
                <label style={{ display: 'block', fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Member Email Address
                </label>
                <div style={{ position: 'relative' }}>
                  <Mail size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.9rem', top: '50%', transform: 'translateY(-50%)' }} />
                  <input
                    type="email"
                    required
                    placeholder="colleague@company.com"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                    style={{ width: '100%', paddingLeft: '2.5rem' }}
                    disabled={isSubmitting}
                  />
                </div>
              </div>

              <div style={{ marginBottom: '1.25rem' }}>
                <label style={{ display: 'block', fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Full Name (Optional)
                </label>
                <div style={{ position: 'relative' }}>
                  <UserIcon size={16} color="var(--text-muted)" style={{ position: 'absolute', left: '0.9rem', top: '50%', transform: 'translateY(-50%)' }} />
                  <input
                    type="text"
                    placeholder="Alex Morgan"
                    value={newFullName}
                    onChange={(e) => setNewFullName(e.target.value)}
                    style={{ width: '100%', paddingLeft: '2.5rem' }}
                    disabled={isSubmitting}
                  />
                </div>
              </div>

              <div style={{ marginBottom: '1.75rem' }}>
                <label style={{ display: 'block', fontSize: '0.82rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Organisation Role
                </label>
                <select
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as any)}
                  style={{ width: '100%' }}
                  disabled={isSubmitting}
                >
                  <option value="member">Member — Can upload, view, and ask questions</option>
                  <option value="admin">Admin — Can also manage team members and edit metadata</option>
                  <option value="owner">Owner — Full administrative control</option>
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                <button type="button" onClick={() => setIsAddModalOpen(false)} className="btn-secondary" disabled={isSubmitting}>
                  Cancel
                </button>
                <button type="submit" className="btn-primary" disabled={isSubmitting}>
                  {isSubmitting ? (
                    <>
                      <Loader2 size={16} className="spinner" />
                      <span>Adding...</span>
                    </>
                  ) : (
                    <>
                      <UserPlus size={16} />
                      <span>Add Member</span>
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
