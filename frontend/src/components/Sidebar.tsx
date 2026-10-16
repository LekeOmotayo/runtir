import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import {
  LayoutDashboard,
  FileText,
  Search,
  Sparkles,
  Users,
  LogOut,
  ChevronDown,
  Building2,
  Check,
  Shield,
} from 'lucide-react';

interface SidebarProps {
  currentTab: string;
  onNavigate: (tab: string) => void;
  onOpenUpload: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ currentTab, onNavigate }) => {
  const { user, activeOrg, organisations, switchOrg, logout } = useAuth();
  const [orgDropdownOpen, setOrgDropdownOpen] = useState(false);

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { id: 'documents', label: 'Documents', icon: FileText },
    { id: 'search', label: 'Content Search', icon: Search },
    { id: 'chat', label: 'AI Q&A Assistant', icon: Sparkles, badge: 'RAG' },
    { id: 'team', label: 'Team & Organisation', icon: Users },
  ];

  return (
    <aside
      style={{
        width: '280px',
        backgroundColor: 'var(--bg-surface)',
        borderRight: '1px solid var(--border-card)',
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      {/* Brand Logo */}
      <div
        style={{
          padding: '1.5rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          borderBottom: '1px solid var(--border-subtle)',
        }}
      >
        <div
          style={{
            width: '38px',
            height: '38px',
            borderRadius: '10px',
            background: 'var(--brand-gradient)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            boxShadow: 'var(--shadow-glow)',
          }}
        >
          <Sparkles size={20} />
        </div>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <span
              style={{
                fontFamily: 'var(--font-display)',
                fontWeight: 800,
                fontSize: '1.25rem',
                letterSpacing: '-0.03em',
                background: 'linear-gradient(to right, #ffffff, #cbd5e1)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              AfriOps
            </span>
            <span
              style={{
                fontSize: '0.65rem',
                fontWeight: 700,
                padding: '0.15rem 0.4rem',
                background: 'rgba(99, 102, 241, 0.15)',
                color: 'var(--brand-light)',
                borderRadius: '4px',
                border: '1px solid rgba(99, 102, 241, 0.3)',
              }}
            >
              MVP
            </span>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            AI Business Operations
          </p>
        </div>
      </div>

      {/* Organisation Switcher */}
      <div style={{ padding: '1rem 1.25rem', position: 'relative' }}>
        <button
          onClick={() => setOrgDropdownOpen(!orgDropdownOpen)}
          style={{
            width: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0.65rem 0.85rem',
            backgroundColor: 'var(--bg-glass-subtle)',
            border: '1px solid var(--border-card)',
            borderRadius: 'var(--radius-md)',
            color: 'var(--text-main)',
            textAlign: 'left',
            transition: 'all 0.2s ease',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', overflow: 'hidden' }}>
            <Building2 size={16} color="var(--brand-light)" />
            <div style={{ overflow: 'hidden' }}>
              <div
                style={{
                  fontSize: '0.85rem',
                  fontWeight: 600,
                  whiteSpace: 'nowrap',
                  textOverflow: 'ellipsis',
                  overflow: 'hidden',
                }}
              >
                {activeOrg?.name || 'My Organisation'}
              </div>
              <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>
                Role: {activeOrg?.role || 'owner'}
              </div>
            </div>
          </div>
          <ChevronDown
            size={14}
            color="var(--text-secondary)"
            style={{
              transform: orgDropdownOpen ? 'rotate(180deg)' : 'none',
              transition: 'transform 0.2s ease',
            }}
          />
        </button>

        {/* Dropdown Menu */}
        {orgDropdownOpen && (
          <div
            className="glass-panel fade-in"
            style={{
              position: 'absolute',
              top: 'calc(100% - 0.25rem)',
              left: '1.25rem',
              right: '1.25rem',
              zIndex: 100,
              padding: '0.5rem',
              boxShadow: 'var(--shadow-lg)',
            }}
          >
            <div
              style={{
                padding: '0.35rem 0.65rem',
                fontSize: '0.72rem',
                color: 'var(--text-muted)',
                fontWeight: 600,
                textTransform: 'uppercase',
              }}
            >
              Your Organisations
            </div>
            {organisations.map((org) => {
              const isActive = org.id === activeOrg?.id;
              return (
                <button
                  key={org.id}
                  onClick={() => {
                    if (!isActive) switchOrg(org.id);
                    setOrgDropdownOpen(false);
                  }}
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0.55rem 0.65rem',
                    borderRadius: 'var(--radius-sm)',
                    backgroundColor: isActive ? 'rgba(99, 102, 241, 0.15)' : 'transparent',
                    color: isActive ? 'var(--brand-light)' : 'var(--text-secondary)',
                    fontSize: '0.84rem',
                    fontWeight: isActive ? 600 : 400,
                    textAlign: 'left',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Building2 size={14} />
                    <span>{org.name}</span>
                  </div>
                  {isActive && <Check size={14} />}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* Navigation Menu */}
      <nav style={{ flex: 1, padding: '0.5rem 1rem', display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0.65rem 0.85rem',
                borderRadius: 'var(--radius-md)',
                backgroundColor: isActive ? 'rgba(99, 102, 241, 0.14)' : 'transparent',
                color: isActive ? 'var(--text-main)' : 'var(--text-secondary)',
                border: isActive ? '1px solid rgba(99, 102, 241, 0.28)' : '1px solid transparent',
                fontSize: '0.9rem',
                fontWeight: isActive ? 600 : 500,
                transition: 'all 0.15s ease',
              }}
              onMouseEnter={(e) => {
                if (!isActive) {
                  e.currentTarget.style.backgroundColor = 'var(--bg-glass-subtle)';
                  e.currentTarget.style.color = 'var(--text-main)';
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive) {
                  e.currentTarget.style.backgroundColor = 'transparent';
                  e.currentTarget.style.color = 'var(--text-secondary)';
                }
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <Icon size={18} color={isActive ? 'var(--brand-light)' : 'currentColor'} />
                <span>{item.label}</span>
              </div>
              {item.badge && (
                <span
                  style={{
                    fontSize: '0.65rem',
                    fontWeight: 700,
                    padding: '0.15rem 0.45rem',
                    borderRadius: '999px',
                    background: 'var(--brand-gradient)',
                    color: 'white',
                  }}
                >
                  {item.badge}
                </span>
              )}
            </button>
          );
        })}
      </nav>

      {/* User Session Footer */}
      <div
        style={{
          padding: '1.25rem',
          borderTop: '1px solid var(--border-subtle)',
          backgroundColor: 'rgba(0, 0, 0, 0.2)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', overflow: 'hidden' }}>
          <div
            style={{
              width: '36px',
              height: '36px',
              borderRadius: '50%',
              backgroundColor: 'var(--bg-elevated)',
              border: '1px solid var(--border-card)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 700,
              color: 'var(--brand-light)',
              fontSize: '0.9rem',
            }}
          >
            {user?.full_name ? user.full_name[0].toUpperCase() : 'U'}
          </div>
          <div style={{ overflow: 'hidden' }}>
            <div
              style={{
                fontSize: '0.85rem',
                fontWeight: 600,
                whiteSpace: 'nowrap',
                textOverflow: 'ellipsis',
                overflow: 'hidden',
              }}
            >
              {user?.full_name}
            </div>
            <div
              style={{
                fontSize: '0.72rem',
                color: 'var(--text-muted)',
                whiteSpace: 'nowrap',
                textOverflow: 'ellipsis',
                overflow: 'hidden',
              }}
            >
              {user?.email}
            </div>
          </div>
        </div>

        <button
          onClick={logout}
          title="Sign out"
          style={{
            padding: '0.5rem',
            color: 'var(--text-secondary)',
            borderRadius: 'var(--radius-sm)',
            transition: 'color 0.2s ease',
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--danger)')}
          onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-secondary)')}
        >
          <LogOut size={18} />
        </button>
      </div>
    </aside>
  );
};
