import React, { useState } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Sidebar } from './components/Sidebar';
import { Header } from './components/Header';
import { DocumentUploadModal } from './components/DocumentUploadModal';
import { DocumentEditModal } from './components/DocumentEditModal';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Dashboard } from './pages/Dashboard';
import { Documents } from './pages/Documents';
import { DocumentDetails } from './pages/DocumentDetails';
import { ContentSearch } from './pages/ContentSearch';
import { AIChat } from './pages/AIChat';
import { TeamSettings } from './pages/TeamSettings';
import { Document } from './types';
import { Loader2, Sparkles } from 'lucide-react';
const MainApp: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();
  const [authView, setAuthView] = useState<'login' | 'register'>('login');

  // App Navigation
  const [currentTab, setCurrentTab] = useState<string>('dashboard');
  const [selectedDoc, setSelectedDoc] = useState<Document | null>(null);
  const [editingDoc, setEditingDoc] = useState<Document | null>(null);
  const [isUploadOpen, setIsUploadOpen] = useState<boolean>(false);
  const [chatInitialPrompt, setChatInitialPrompt] = useState<string | undefined>();
  const [refreshKey, setRefreshKey] = useState<number>(0);

  if (isLoading) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: 'var(--bg-base)',
          gap: '1rem',
        }}
      >
        <div
          style={{
            width: '48px',
            height: '48px',
            borderRadius: '12px',
            background: 'var(--brand-gradient)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            boxShadow: 'var(--brand-glow)',
          }}
        >
          <Sparkles size={24} />
        </div>
        <Loader2 size={24} className="spinner" color="var(--brand-light)" />
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>
          Initializing AfriOps environment...
        </p>
      </div>
    );
  }

  // Not authenticated: render Login / Register
  if (!isAuthenticated) {
    if (authView === 'register') {
      return <Register onGoToLogin={() => setAuthView('login')} />;
    }
    return <Login onGoToRegister={() => setAuthView('register')} />;
  }

  // Handle navigation
  const handleNavigate = (tab: string, docID?: string) => {
    setCurrentTab(tab);
    if (tab !== 'documents') {
      setSelectedDoc(null);
    }
  };

  const handleSelectDoc = (doc: Document) => {
    setSelectedDoc(doc);
    setCurrentTab('documents');
  };

  const handleAskAIAboutPassage = (passage: string, docTitle: string) => {
    setChatInitialPrompt(`Regarding "${docTitle}", can you explain the following passage: "${passage}"`);
    setCurrentTab('chat');
  };

  // Titles mapping
  const getHeaderInfo = () => {
    if (selectedDoc) {
      return {
        title: selectedDoc.title,
        subtitle: `Document details, metadata, and extraction pipeline status`,
      };
    }
    switch (currentTab) {
      case 'dashboard':
        return {
          title: 'Operations Dashboard',
          subtitle: 'Real-time overview of document ingestion, processing statuses, and operational metrics',
        };
      case 'documents':
        return {
          title: 'Document Repository',
          subtitle: 'Search, filter, view, archive, and manage your organisation files',
        };
      case 'search':
        return {
          title: 'Passage & Content Search',
          subtitle: 'Search inside document text extractions with relevant passage snippets',
        };
      case 'chat':
        return {
          title: 'AI Q&A Assistant',
          subtitle: 'Ask natural-language questions grounded in verified source citations',
        };
      case 'team':
        return {
          title: 'Team & Organisation Settings',
          subtitle: 'Manage organisation profile, member access roles, and permissions',
        };
      default:
        return { title: 'AfriOps', subtitle: 'AI Operations' };
    }
  };

  const { title, subtitle } = getHeaderInfo();

  return (
    <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: 'var(--bg-base)' }}>
      {/* Left Sidebar */}
      <Sidebar
        currentTab={currentTab}
        onNavigate={(tab) => {
          setSelectedDoc(null);
          setCurrentTab(tab);
        }}
        onOpenUpload={() => setIsUploadOpen(true)}
      />

      {/* Main Content Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <Header
          title={title}
          subtitle={subtitle}
          onOpenUpload={() => setIsUploadOpen(true)}
          onSearchClick={() => {
            setSelectedDoc(null);
            setCurrentTab('search');
          }}
        />

        <main style={{ flex: 1, overflowY: 'auto' }}>
          {selectedDoc ? (
            <DocumentDetails
              document={selectedDoc}
              onBack={() => setSelectedDoc(null)}
              onEdit={(doc) => setEditingDoc(doc)}
              onRefreshList={() => setRefreshKey((k) => k + 1)}
            />
          ) : (
            <>
              {currentTab === 'dashboard' && (
                <Dashboard
                  onOpenUpload={() => setIsUploadOpen(true)}
                  onNavigate={(tab, docID) => {
                    if (docID) {
                      // Fetch or navigate to doc
                      setCurrentTab('documents');
                    } else {
                      setCurrentTab(tab);
                    }
                  }}
                />
              )}

              {currentTab === 'documents' && (
                <Documents
                  key={refreshKey}
                  onOpenUpload={() => setIsUploadOpen(true)}
                  onSelectDocument={handleSelectDoc}
                  onEditDocument={(doc) => setEditingDoc(doc)}
                />
              )}

              {currentTab === 'search' && (
                <ContentSearch
                  onAskAIAboutPassage={handleAskAIAboutPassage}
                  onNavigateToDoc={(docID) => {
                    // Navigate to documents list
                    setCurrentTab('documents');
                  }}
                />
              )}

              {currentTab === 'chat' && (
                <AIChat initialQuestion={chatInitialPrompt} />
              )}

              {currentTab === 'team' && (
                <TeamSettings />
              )}
            </>
          )}
        </main>
      </div>

      {/* Global Document Upload Modal */}
      <DocumentUploadModal
        isOpen={isUploadOpen}
        onClose={() => setIsUploadOpen(false)}
        onSuccess={(newDoc) => {
          setRefreshKey((k) => k + 1);
          handleSelectDoc(newDoc);
        }}
      />

      {/* Global Document Edit Modal */}
      <DocumentEditModal
        document={editingDoc}
        isOpen={!!editingDoc}
        onClose={() => setEditingDoc(null)}
        onSuccess={(updated) => {
          setRefreshKey((k) => k + 1);
          if (selectedDoc?.id === updated.id) {
            setSelectedDoc(updated);
          }
        }}
      />
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
};
export default App;
