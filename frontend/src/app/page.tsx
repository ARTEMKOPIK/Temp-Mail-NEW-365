'use client';

import { useState, useEffect, useCallback } from 'react';
import { Copy, RefreshCw, Mail, Clock, Shield, Zap, Eye } from 'lucide-react';
import { createMailbox, getEmails, getEmail, type Mailbox, type EmailListItem, type Email } from '@/lib/api';
import { connectToMailbox, type MailboxWebSocket, type WebSocketMessage } from '@/lib/websocket';

export default function Home() {
  const [mailbox, setMailbox] = useState<Mailbox | null>(null);
  const [emails, setEmails] = useState<EmailListItem[]>([]);
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
  const [loading, setLoading] = useState(false);
  const [expiryMinutes, setExpiryMinutes] = useState(60);
  const [timeLeft, setTimeLeft] = useState<string>('');
  const [wsConnection, setWsConnection] = useState<MailboxWebSocket | null>(null);
  const [copySuccess, setCopySuccess] = useState(false);

  // Calculate time left
  useEffect(() => {
    if (!mailbox) return;

    const updateTimeLeft = () => {
      const now = new Date();
      const expires = new Date(mailbox.expiresAt);
      const diff = expires.getTime() - now.getTime();

      if (diff <= 0) {
        setTimeLeft('Expired');
        return;
      }

      const minutes = Math.floor(diff / 60000);
      const seconds = Math.floor((diff % 60000) / 1000);
      setTimeLeft(`${minutes}m ${seconds}s`);
    };

    updateTimeLeft();
    const interval = setInterval(updateTimeLeft, 1000);

    return () => clearInterval(interval);
  }, [mailbox]);

  // Load emails
  const loadEmails = useCallback(async (address: string) => {
    try {
      const emailList = await getEmails(address);
      setEmails(emailList);
    } catch (error) {
      console.error('Failed to load emails:', error);
    }
  }, []);

  // WebSocket connection
  useEffect(() => {
    if (!mailbox) return;

    const ws = connectToMailbox(mailbox.address);
    setWsConnection(ws);

    const unsubscribe = ws.onMessage((message: WebSocketMessage) => {
      if (message.type === 'new_email') {
        console.log('New email received!');
        loadEmails(mailbox.address);
      }
    });

    // Load initial emails
    loadEmails(mailbox.address);

    return () => {
      unsubscribe();
      ws.disconnect();
    };
  }, [mailbox, loadEmails]);

  // Create new mailbox
  const handleCreateMailbox = async () => {
    setLoading(true);
    try {
      const newMailbox = await createMailbox(expiryMinutes);
      setMailbox(newMailbox);
      setEmails([]);
      setSelectedEmail(null);
    } catch (error) {
      console.error('Failed to create mailbox:', error);
      alert('Failed to create mailbox. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Copy address to clipboard
  const handleCopyAddress = async () => {
    if (!mailbox) return;
    
    try {
      await navigator.clipboard.writeText(mailbox.address);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  // View email
  const handleViewEmail = async (emailId: string) => {
    try {
      const email = await getEmail(emailId);
      setSelectedEmail(email);
    } catch (error) {
      console.error('Failed to load email:', error);
      alert('Failed to load email. Please try again.');
    }
  };

  // Initial load
  useEffect(() => {
    handleCreateMailbox();
  }, []);

  return (
    <main className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Mail className="w-8 h-8 text-blue-600" />
              <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                Temp Mail 365
              </h1>
            </div>
            <nav className="flex gap-6">
              <a href="#features" className="text-gray-600 hover:text-gray-900">Features</a>
              <a href="/faq" className="text-gray-600 hover:text-gray-900">FAQ</a>
              <a href="/blog" className="text-gray-600 hover:text-gray-900">Blog</a>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 py-12">
        <div className="text-center mb-12 animate-fade-in">
          <h2 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            Temporary Email Address
          </h2>
          <p className="text-xl text-gray-600 mb-8">
            Protect your privacy with disposable email addresses
          </p>
        </div>

        {/* Mailbox Card */}
        <div className="max-w-4xl mx-auto mb-12">
          <div className="bg-white rounded-2xl shadow-xl p-8 animate-slide-up">
            {/* Settings */}
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-4">
                <label className="text-sm font-medium text-gray-700">Lifetime:</label>
                <select
                  value={expiryMinutes}
                  onChange={(e) => setExpiryMinutes(Number(e.target.value))}
                  className="px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                  disabled={!!mailbox}
                >
                  <option value={5}>5 minutes</option>
                  <option value={10}>10 minutes</option>
                  <option value={30}>30 minutes</option>
                  <option value={60}>60 minutes</option>
                  <option value={120}>2 hours</option>
                </select>
              </div>
              
              <button
                onClick={handleCreateMailbox}
                disabled={loading}
                className="flex items-center gap-2 px-6 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:shadow-lg transition-all disabled:opacity-50"
              >
                <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                New Address
              </button>
            </div>

            {/* Email Address Display */}
            {mailbox && (
              <div className="space-y-4">
                <div className="flex items-center gap-4 p-4 bg-gray-50 rounded-lg">
                  <input
                    type="text"
                    value={mailbox.address}
                    readOnly
                    className="flex-1 bg-transparent text-lg font-mono outline-none"
                  />
                  <button
                    onClick={handleCopyAddress}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                    {copySuccess ? 'Copied!' : 'Copy'}
                  </button>
                </div>

                {/* Timer */}
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2 text-gray-600">
                    <Clock className="w-4 h-4" />
                    <span>Expires in: <strong className="text-blue-600">{timeLeft}</strong></span>
                  </div>
                  <div className="text-gray-600">
                    <span>Emails: <strong className="text-purple-600">{emails.length}</strong></span>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Inbox */}
        {mailbox && (
          <div className="max-w-4xl mx-auto grid md:grid-cols-2 gap-6">
            {/* Email List */}
            <div className="bg-white rounded-2xl shadow-xl p-6">
              <h3 className="text-xl font-bold mb-4 flex items-center gap-2">
                <Mail className="w-5 h-5 text-blue-600" />
                Inbox ({emails.length})
              </h3>
              
              {emails.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                  <Mail className="w-16 h-16 mx-auto mb-4 opacity-20" />
                  <p>No emails yet...</p>
                  <p className="text-sm">Waiting for incoming mail</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {emails.map((email) => (
                    <button
                      key={email.id}
                      onClick={() => handleViewEmail(email.id)}
                      className="w-full text-left p-4 rounded-lg hover:bg-gray-50 border border-gray-100 transition-colors"
                    >
                      <div className="font-semibold text-gray-900 truncate">{email.subject || '(No Subject)'}</div>
                      <div className="text-sm text-gray-600 truncate">{email.from}</div>
                      <div className="text-xs text-gray-400 mt-1">
                        {new Date(email.receivedAt).toLocaleString()}
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Email Viewer */}
            <div className="bg-white rounded-2xl shadow-xl p-6">
              <h3 className="text-xl font-bold mb-4 flex items-center gap-2">
                <Eye className="w-5 h-5 text-purple-600" />
                Email Content
              </h3>
              
              {!selectedEmail ? (
                <div className="text-center py-12 text-gray-400">
                  <Eye className="w-16 h-16 mx-auto mb-4 opacity-20" />
                  <p>Select an email to view</p>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="border-b pb-4">
                    <div className="font-semibold text-lg mb-2">{selectedEmail.subject || '(No Subject)'}</div>
                    <div className="text-sm text-gray-600">From: {selectedEmail.from}</div>
                    <div className="text-xs text-gray-400 mt-1">
                      {new Date(selectedEmail.receivedAt).toLocaleString()}
                    </div>
                  </div>
                  
                  <div className="prose max-w-none">
                    {selectedEmail.htmlBody ? (
                      <div dangerouslySetInnerHTML={{ __html: selectedEmail.htmlBody }} />
                    ) : (
                      <pre className="whitespace-pre-wrap font-sans text-sm">{selectedEmail.textBody}</pre>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </section>

      {/* Features Section */}
      <section id="features" className="container mx-auto px-4 py-16">
        <h2 className="text-3xl font-bold text-center mb-12">Why Choose Temp Mail 365?</h2>
        
        <div className="grid md:grid-cols-3 gap-8">
          <div className="bg-white p-8 rounded-2xl shadow-lg text-center">
            <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Shield className="w-8 h-8 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">Privacy Protected</h3>
            <p className="text-gray-600">No registration, no personal data stored. Complete anonymity.</p>
          </div>

          <div className="bg-white p-8 rounded-2xl shadow-lg text-center">
            <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Zap className="w-8 h-8 text-purple-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">Instant Access</h3>
            <p className="text-gray-600">Get your temporary email in seconds. No waiting, no hassle.</p>
          </div>

          <div className="bg-white p-8 rounded-2xl shadow-lg text-center">
            <div className="w-16 h-16 bg-pink-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Mail className="w-8 h-8 text-pink-600" />
            </div>
            <h3 className="text-xl font-bold mb-2">Real-time Updates</h3>
            <p className="text-gray-600">Receive emails instantly with WebSocket notifications.</p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 py-12 mt-16">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-4 gap-8">
            <div>
              <h4 className="font-bold text-white mb-4">Temp Mail 365</h4>
              <p className="text-sm">Secure temporary email service</p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-4">Product</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="/tutorials" className="hover:text-white">Tutorials</a></li>
                <li><a href="/faq" className="hover:text-white">FAQ</a></li>
                <li><a href="/blog" className="hover:text-white">Blog</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-bold text-white mb-4">Legal</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="/privacy" className="hover:text-white">Privacy Policy</a></li>
                <li><a href="/terms" className="hover:text-white">Terms of Service</a></li>
              </ul>
            </div>
            <div>
              <h4 className="font-bold text-white mb-4">Support</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="/contact" className="hover:text-white">Contact Us</a></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-sm">
            © 2025 Temp Mail 365. All rights reserved.
          </div>
        </div>
      </footer>
    </main>
  );
}

