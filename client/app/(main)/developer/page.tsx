'use client';

import { useState, useEffect, useCallback } from 'react';
import { Icon } from '@iconify/react';
import { toast } from 'sonner';
import {
  developerService,
  type APIApplication,
  type APIKey,
  type GeneratedKeyPair,
  type RequestLog,
  type Webhook,
} from '@/feature/developer/developer.service';

type Tab = 'apps' | 'detail';
type DetailTab = 'keys' | 'logs' | 'webhooks';

const WEBHOOK_EVENTS = [
  'deposit.confirmed',
  'withdrawal.completed',
  'trade.created',
  'trade.paid',
  'trade.released',
  'trade.cancelled',
];

export default function DeveloperPage() {
  const [activeTab, setActiveTab] = useState<Tab>('apps');
  const [apps, setApps] = useState<APIApplication[]>([]);
  const [selectedApp, setSelectedApp] = useState<APIApplication | null>(null);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newAppName, setNewAppName] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newKeys, setNewKeys] = useState<GeneratedKeyPair | null>(null);
  const [copiedField, setCopiedField] = useState<string | null>(null);

  // Detail state
  const [detailTab, setDetailTab] = useState<DetailTab>('keys');
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [logs, setLogs] = useState<RequestLog[]>([]);
  const [logsTotal, setLogsTotal] = useState(0);
  const [logsPage, setLogsPage] = useState(1);
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [showWebhookForm, setShowWebhookForm] = useState(false);
  const [webhookUrl, setWebhookUrl] = useState('');
  const [webhookEvents, setWebhookEvents] = useState<string[]>([]);
  const [newWebhookSecret, setNewWebhookSecret] = useState<string | null>(null);
  const [showRegenerateModal, setShowRegenerateModal] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const [openDropdownId, setOpenDropdownId] = useState<string | null>(null);

  const fetchApps = useCallback(async () => {
    try {
      setLoading(true);
      const data = await developerService.listApps();
      setApps(data);
    } catch (err) {
      console.error('Failed to load apps:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchApps();
  }, [fetchApps]);

  const handleCreateApp = async () => {
    if (!newAppName.trim()) return;
    setCreating(true);
    try {
      const result = await developerService.createApp(newAppName.trim());
      setNewKeys(result.keys);
      setNewAppName('');
      setShowCreateForm(false);
      await fetchApps();
    } catch (err) {
      console.error('Failed to create app:', err);
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteApp = (appId: string) => {
    toast('Delete this app?', {
      description: 'All keys will be revoked.',
      action: {
        label: 'Delete',
        onClick: async () => {
          try {
            await developerService.deleteApp(appId);
            if (selectedApp?.id === appId) {
              setSelectedApp(null);
              setActiveTab('apps');
            }
            await fetchApps();
            toast.success('App deleted successfully');
          } catch (err) {
            console.error('Failed to delete app:', err);
            toast.error('Failed to delete app');
          }
        }
      },
      cancel: { label: 'Cancel', onClick: () => {} }
    });
  };

  const handleToggleLiveStatus = async (app: APIApplication) => {
    try {
      const newStatus = !app.is_live;
      await developerService.updateAppStatus(app.id, newStatus);
      toast.success(`Application is now in ${newStatus ? 'Live' : 'Test'} Mode`);
      
      const updatedApp = { ...app, is_live: newStatus };
      if (selectedApp?.id === app.id) {
        setSelectedApp(updatedApp);
      }
      setApps(prev => prev.map(a => a.id === app.id ? updatedApp : a));
    } catch (err: any) {
      console.error('Failed to update app status:', err);
      toast.error(err?.response?.data?.message || 'Failed to update app status');
    }
  };

  const openAppDetail = async (app: APIApplication) => {
    setSelectedApp(app);
    setActiveTab('detail');
    setDetailTab('keys');
    await loadKeys(app.id);
  };

  const loadKeys = async (appId: string) => {
    try {
      const data = await developerService.listKeys(appId);
      setKeys(data);
    } catch (err) {
      console.error('Failed to load keys:', err);
    }
  };

  const loadLogs = async (appId: string, page = 1) => {
    try {
      const data = await developerService.getLogs(appId, page);
      setLogs(data.logs || []);
      setLogsTotal(data.total);
      setLogsPage(data.page);
    } catch (err) {
      console.error('Failed to load logs:', err);
    }
  };

  const loadWebhooks = async (appId: string) => {
    try {
      const data = await developerService.listWebhooks(appId);
      setWebhooks(data);
    } catch (err) {
      console.error('Failed to load webhooks:', err);
    }
  };

  const handleRegenerateKeys = () => {
    if (!selectedApp) return;
    setShowRegenerateModal(true);
  };

  const confirmRegenerateKeys = async () => {
    if (!selectedApp) return;
    setRegenerating(true);
    try {
      const result = await developerService.regenerateKeys(selectedApp.id);
      setNewKeys(result.keys);
      await loadKeys(selectedApp.id);
      toast.success(result.message || 'Keys regenerated successfully');
      setShowRegenerateModal(false);
    } catch (err: any) {
      console.error('Failed to regenerate keys:', err);
      toast.error(err?.response?.data?.message || 'Failed to regenerate keys');
    } finally {
      setRegenerating(false);
    }
  };

  const handleRevokeKey = (keyId: string) => {
    if (!selectedApp) return;
    toast('Revoke this key?', {
      description: 'This action cannot be undone.',
      action: {
        label: 'Revoke',
        onClick: async () => {
          try {
            await developerService.revokeKey(selectedApp.id, keyId);
            await loadKeys(selectedApp.id);
            toast.success('Key revoked successfully');
          } catch (err) {
            console.error('Failed to revoke key:', err);
            toast.error('Failed to revoke key');
          }
        }
      },
      cancel: { label: 'Cancel', onClick: () => {} }
    });
  };

  const handleCreateWebhook = async () => {
    if (!selectedApp || !webhookUrl.trim()) return;
    try {
      const result = await developerService.createWebhook(selectedApp.id, webhookUrl.trim(), webhookEvents);
      setNewWebhookSecret(result.webhook.secret || null);
      setWebhookUrl('');
      setWebhookEvents([]);
      setShowWebhookForm(false);
      await loadWebhooks(selectedApp.id);
    } catch (err) {
      console.error('Failed to create webhook:', err);
    }
  };

  const handleDeleteWebhook = (webhookId: string) => {
    if (!selectedApp) return;
    toast('Delete this webhook?', {
      description: 'You will stop receiving events at this URL.',
      action: {
        label: 'Delete',
        onClick: async () => {
          try {
            await developerService.deleteWebhook(selectedApp.id, webhookId);
            await loadWebhooks(selectedApp.id);
            toast.success('Webhook deleted successfully');
          } catch (err) {
            console.error('Failed to delete webhook:', err);
            toast.error('Failed to delete webhook');
          }
        }
      },
      cancel: { label: 'Cancel', onClick: () => {} }
    });
  };

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const toggleWebhookEvent = (event: string) => {
    setWebhookEvents(prev =>
      prev.includes(event) ? prev.filter(e => e !== event) : [...prev, event]
    );
  };

  const getStatusColor = (code: number) => {
    if (code < 300) return 'text-emerald-400';
    if (code < 400) return 'text-amber-400';
    return 'text-red-400';
  };

  // ── Loading state ─────────────────────────────────────────────────────────
  if (loading) {
    return (
      <div className="flex justify-center pt-20">
        <div className="animate-spin text-[#4f7cf7]">
          <Icon icon="hugeicons:loading-03" className="size-8" />
        </div>
      </div>
    );
  }

  // ── New keys modal ────────────────────────────────────────────────────────
  const renderKeysModal = () => {
    if (!newKeys) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80">
        <div className="w-full max-w-[520px] bg-[#13151a] border border-[#22252e] rounded-[24px] p-8 shadow-2xl flex flex-col gap-6">
          <div className="flex items-center gap-3">
            <div className="size-12 rounded-2xl bg-amber-500/10 flex items-center justify-center">
              <Icon icon="hugeicons:alert-02" className="size-6 text-amber-400" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">Save Your API Keys</h2>
              <p className="text-sm text-[#888c99]">These keys will only be shown once.</p>
            </div>
          </div>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold text-[#888c99] uppercase tracking-wider">Secret Key</label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl text-sm text-red-300 font-mono truncate">
                  {newKeys.secret_key}
                </code>
                <button
                  onClick={() => copyToClipboard(newKeys.secret_key, 'sk')}
                  className="size-10 shrink-0 bg-[#1c1f26] hover:bg-[#22252e] border border-[#22252e] rounded-xl flex items-center justify-center transition-colors"
                >
                  <Icon icon={copiedField === 'sk' ? 'hugeicons:checkmark-circle-02' : 'hugeicons:copy-01'} className="size-4 text-white" />
                </button>
              </div>
              <p className="text-[10px] text-red-400/80">⚠ Keep this secret. Only use in your backend server.</p>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-xs font-semibold text-[#888c99] uppercase tracking-wider">Publishable Key</label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl text-sm text-emerald-300 font-mono truncate">
                  {newKeys.publishable_key}
                </code>
                <button
                  onClick={() => copyToClipboard(newKeys.publishable_key, 'pk')}
                  className="size-10 shrink-0 bg-[#1c1f26] hover:bg-[#22252e] border border-[#22252e] rounded-xl flex items-center justify-center transition-colors"
                >
                  <Icon icon={copiedField === 'pk' ? 'hugeicons:checkmark-circle-02' : 'hugeicons:copy-01'} className="size-4 text-white" />
                </button>
              </div>
              <p className="text-[10px] text-[#888c99]">Safe to use in your frontend.</p>
            </div>
          </div>

          <button
            onClick={() => setNewKeys(null)}
            className="w-full py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition-colors shadow-lg shadow-blue-600/20"
          >
            I've saved my keys
          </button>
        </div>
      </div>
    );
  };

  // ── Webhook secret modal ──────────────────────────────────────────────────
  const renderWebhookSecretModal = () => {
    if (!newWebhookSecret) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80">
        <div className="w-full max-w-[480px] bg-[#13151a] border border-[#22252e] rounded-[24px] p-8 shadow-2xl flex flex-col gap-6">
          <div className="flex items-center gap-3">
            <div className="size-12 rounded-2xl bg-amber-500/10 flex items-center justify-center">
              <Icon icon="hugeicons:key-01" className="size-6 text-amber-400" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">Webhook Secret</h2>
              <p className="text-sm text-[#888c99]">Use this to verify webhook signatures.</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <code className="flex-1 px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl text-sm text-amber-300 font-mono truncate">
              {newWebhookSecret}
            </code>
            <button
              onClick={() => copyToClipboard(newWebhookSecret, 'wh')}
              className="size-10 shrink-0 bg-[#1c1f26] hover:bg-[#22252e] border border-[#22252e] rounded-xl flex items-center justify-center transition-colors"
            >
              <Icon icon={copiedField === 'wh' ? 'hugeicons:checkmark-circle-02' : 'hugeicons:copy-01'} className="size-4 text-white" />
            </button>
          </div>
          <button
            onClick={() => setNewWebhookSecret(null)}
            className="w-full py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition-colors"
          >
            Done
          </button>
        </div>
      </div>
    );
  };

  // ── Create App modal ──────────────────────────────────────────────────────
  const renderCreateAppModal = () => {
    if (!showCreateForm) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 animate-in fade-in duration-200">
        <div className="w-full max-w-[480px] bg-[#13151a] border border-[#22252e] rounded-[24px] p-6 sm:p-8 shadow-2xl flex flex-col gap-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-white">Create Application</h2>
              <p className="text-sm text-[#888c99] mt-1">Generate a new set of API keys.</p>
            </div>
            <button onClick={() => setShowCreateForm(false)} className="text-[#888c99] hover:text-white transition-colors">
              <Icon icon="hugeicons:cancel-01" className="size-6" />
            </button>
          </div>
          <div className="flex flex-col gap-4">
            <input
              type="text"
              value={newAppName}
              onChange={(e) => setNewAppName(e.target.value)}
              placeholder="Application name (e.g. My Crypto Shop)"
              className="w-full px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl text-white text-sm placeholder:text-[#888c99] focus:outline-none focus:border-blue-500 transition-colors"
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreateApp();
              }}
            />
            <div className="flex gap-3 mt-2">
              <button
                onClick={() => setShowCreateForm(false)}
                className="flex-1 py-3 bg-[#1c1f26] hover:bg-[#22252e] text-white font-semibold rounded-xl transition-colors border border-[#22252e]"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateApp}
                disabled={creating || !newAppName.trim()}
                className="flex-1 py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-[#22252e] disabled:text-[#888c99] text-white font-semibold rounded-xl transition-colors"
              >
                {creating ? 'Creating...' : 'Create App'}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // ── Regenerate Keys modal ──────────────────────────────────────────────────
  const renderRegenerateModal = () => {
    if (!showRegenerateModal) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 animate-in fade-in duration-200">
        <div className="w-full max-w-[480px] bg-[#13151a] border border-[#22252e] rounded-[24px] p-6 sm:p-8 shadow-2xl flex flex-col gap-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="size-12 rounded-2xl bg-amber-500/10 flex items-center justify-center">
                <Icon icon="hugeicons:alert-02" className="size-6 text-amber-400" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-white">Regenerate Keys</h2>
              </div>
            </div>
            <button onClick={() => setShowRegenerateModal(false)} className="text-[#888c99] hover:text-white transition-colors">
              <Icon icon="hugeicons:cancel-01" className="size-6" />
            </button>
          </div>
          
          <div className="text-sm text-[#888c99]">
            Are you sure you want to regenerate API keys for this application? All existing keys will be immediately revoked and will stop working. This action cannot be undone.
          </div>

          <div className="flex gap-3 mt-2">
            <button
              onClick={() => setShowRegenerateModal(false)}
              className="flex-1 py-3 bg-[#1c1f26] hover:bg-[#22252e] text-white font-semibold rounded-xl transition-colors border border-[#22252e]"
            >
              Cancel
            </button>
            <button
              onClick={confirmRegenerateKeys}
              disabled={regenerating}
              className="flex-1 py-3 bg-amber-600 hover:bg-amber-500 disabled:bg-[#22252e] disabled:text-[#888c99] text-white font-semibold rounded-xl transition-colors"
            >
              {regenerating ? 'Regenerating...' : 'Yes, Regenerate'}
            </button>
          </div>
        </div>
      </div>
    );
  };

  // ── Webhook modal ─────────────────────────────────────────────────────────
  const renderWebhookModal = () => {
    if (!showWebhookForm) return null;
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 animate-in fade-in duration-200">
        <div className="w-full max-w-[520px] bg-[#13151a] border border-[#22252e] rounded-[24px] p-6 sm:p-8 shadow-2xl flex flex-col gap-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-white">Add Webhook</h2>
              <p className="text-sm text-[#888c99] mt-1">Receive real-time notifications for events.</p>
            </div>
            <button onClick={() => setShowWebhookForm(false)} className="text-[#888c99] hover:text-white transition-colors">
              <Icon icon="hugeicons:cancel-01" className="size-6" />
            </button>
          </div>
          <div className="flex flex-col gap-5">
            <div>
              <label className="text-xs font-semibold text-[#888c99] uppercase tracking-wider mb-2 block">Endpoint URL</label>
              <input
                type="url"
                value={webhookUrl}
                onChange={(e) => setWebhookUrl(e.target.value)}
                placeholder="https://your-app.com/webhook"
                className="w-full px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl text-white text-sm placeholder:text-[#888c99] focus:outline-none focus:border-blue-500 transition-colors"
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-[#888c99] uppercase tracking-wider mb-2 block">Events to listen for</label>
              <div className="flex flex-wrap gap-2">
                {WEBHOOK_EVENTS.map(event => (
                  <button
                    key={event}
                    onClick={() => toggleWebhookEvent(event)}
                    className={`px-3 py-2 rounded-lg text-xs font-semibold transition-colors border ${webhookEvents.includes(event)
                        ? 'bg-blue-600/20 text-blue-400 border-blue-500/30'
                        : 'bg-[#1c1f26] text-[#888c99] border-[#22252e] hover:text-white'
                      }`}
                  >
                    {event}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex gap-3 mt-2">
              <button
                onClick={() => setShowWebhookForm(false)}
                className="flex-1 py-3 bg-[#1c1f26] hover:bg-[#22252e] text-white font-semibold rounded-xl transition-colors border border-[#22252e]"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateWebhook}
                disabled={!webhookUrl.trim()}
                className="flex-1 py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-[#22252e] disabled:text-[#888c99] text-white font-semibold rounded-xl transition-colors"
              >
                Create Webhook
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // ── Apps list ─────────────────────────────────────────────────────────────
  const renderApps = () => (
    <div className="flex flex-col gap-6 animate-in fade-in zoom-in-95 duration-300">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">Your Applications</h2>
          <p className="text-sm text-[#888c99] mt-1">Create apps to get API keys for integration.</p>
        </div>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-colors shadow-lg shadow-blue-600/20"
        >
          <Icon icon="hugeicons:add-circle-half-dot" className="size-4" />
          New App
        </button>
      </div>

      {/* Apps Grid */}
      {apps.length === 0 ? (
        <div className="p-12 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col items-center justify-center gap-4 text-center">
          <div className="size-16 rounded-2xl bg-blue-600/10 flex items-center justify-center">
            <Icon icon="hugeicons:api" className="size-8 text-blue-400" />
          </div>
          <div>
            <p className="text-white font-semibold">No applications yet</p>
            <p className="text-sm text-[#888c99] mt-1">Create your first app to get started with the API.</p>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {apps.map(app => (
            <div
              key={app.id}
              className="p-5 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-3 hover:border-blue-500/30 transition-colors group relative"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded-xl bg-blue-600/10 flex items-center justify-center">
                    <Icon icon="hugeicons:code" className="size-5 text-blue-400" />
                  </div>
                  <div>
                    <h3 className="text-white font-semibold group-hover:text-blue-400 transition-colors">{app.name}</h3>
                    <p className="text-xs text-[#888c99]">
                      Created {new Date(app.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="relative">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setOpenDropdownId(openDropdownId === app.id ? null : app.id);
                    }}
                    className="size-8 rounded-lg hover:bg-[#22252e] flex items-center justify-center transition-colors"
                  >
                    <Icon icon="hugeicons:more-vertical" className="size-5 text-[#888c99]" />
                  </button>
                  {openDropdownId === app.id && (
                    <>
                      <div 
                        className="fixed inset-0 z-40" 
                        onClick={(e) => { e.stopPropagation(); setOpenDropdownId(null); }} 
                      />
                      <div className="absolute right-0 mt-2 w-36 bg-[#1c1f26] border border-[#22252e] rounded-xl shadow-xl z-50 overflow-hidden animate-in fade-in slide-in-from-top-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setOpenDropdownId(null);
                            openAppDetail(app);
                          }}
                          className="w-full px-4 py-3 text-left text-sm text-white hover:bg-[#22252e] flex items-center gap-2 transition-colors"
                        >
                          <Icon icon="hugeicons:view" className="size-4 text-[#888c99]" />
                          View
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setOpenDropdownId(null);
                            handleDeleteApp(app.id);
                          }}
                          className="w-full px-4 py-3 text-left text-sm text-red-400 hover:bg-red-500/10 flex items-center gap-2 transition-colors"
                        >
                          <Icon icon="hugeicons:delete-02" className="size-4" />
                          Delete
                        </button>
                      </div>
                    </>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${app.is_live
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                    : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                  }`}>
                  {app.is_live ? 'Live' : 'Test Mode'}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  // ── App detail ────────────────────────────────────────────────────────────
  const renderDetail = () => {
    if (!selectedApp) return null;

    return (
      <div className="flex flex-col gap-6 animate-in fade-in zoom-in-95 duration-300">
        {/* Back + Title + Toggle */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => { setActiveTab('apps'); setSelectedApp(null); }}
              className="size-10 rounded-xl bg-[#1c1f26] border border-[#22252e] flex items-center justify-center hover:bg-[#22252e] transition-colors"
            >
              <Icon icon="hugeicons:arrow-left-01" className="size-5 text-white" />
            </button>
            <div>
              <h2 className="text-xl font-bold text-white">{selectedApp.name}</h2>
              <p className="text-xs text-[#888c99] font-mono">{selectedApp.id}</p>
            </div>
          </div>
          
          <div className="flex items-center gap-3">
            <span className={`text-xs font-semibold ${selectedApp.is_live ? 'text-emerald-400' : 'text-amber-400'}`}>
              {selectedApp.is_live ? 'Live Mode' : 'Test Mode'}
            </span>
            <button
              onClick={() => handleToggleLiveStatus(selectedApp)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none ${selectedApp.is_live ? 'bg-emerald-500' : 'bg-[#22252e]'}`}
            >
              <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${selectedApp.is_live ? 'translate-x-6' : 'translate-x-1'}`} />
            </button>
          </div>
        </div>

        {/* Detail Tabs */}
        <div className="flex items-center gap-2 border-b border-[#22252e]">
          {(['keys', 'logs', 'webhooks'] as DetailTab[]).map(tab => (
            <button
              key={tab}
              onClick={() => {
                setDetailTab(tab);
                if (tab === 'keys') loadKeys(selectedApp.id);
                if (tab === 'logs') loadLogs(selectedApp.id);
                if (tab === 'webhooks') loadWebhooks(selectedApp.id);
              }}
              className={`pb-4 px-4 text-[15px] font-semibold transition-all relative capitalize ${detailTab === tab ? 'text-white' : 'text-[#888c99] hover:text-white'
                }`}
            >
              {tab}
              {detailTab === tab && (
                <div className="absolute bottom-0 left-0 w-full h-[2px] bg-blue-600 rounded-t-full" />
              )}
            </button>
          ))}
        </div>

        {/* Keys Tab */}
        {detailTab === 'keys' && (
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-[#888c99]">API keys for this application</p>
              <button
                onClick={handleRegenerateKeys}
                className="flex items-center gap-2 px-4 py-2 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 text-sm font-semibold rounded-xl transition-colors border border-amber-500/20"
              >
                <Icon icon="hugeicons:refresh" className="size-4" />
                Regenerate
              </button>
            </div>

            {keys.length === 0 ? (
              <div className="p-8 bg-[#1c1f26] border border-[#22252e] rounded-2xl text-center text-[#888c99]">
                No keys found. Click Regenerate to create new ones.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {keys.map(key => (
                  <div key={key.id} className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex items-start justify-between">
                    <div className="flex items-start gap-4">
                      <div className={`size-10 rounded-xl flex items-center justify-center shrink-0 ${key.is_active ? 'bg-blue-600/10' : 'bg-red-500/10'
                        }`}>
                        <Icon
                          icon="hugeicons:key-01"
                          className={`size-5 ${key.is_active ? 'text-blue-400' : 'text-red-400'}`}
                        />
                      </div>
                      <div className="flex flex-col gap-2">
                        <div className="flex items-center gap-2">
                          <span className="text-white font-semibold">API Key Pair</span>
                          <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${key.is_active
                              ? 'bg-emerald-500/10 text-emerald-400'
                              : 'bg-red-500/10 text-red-400'
                            }`}>
                            {key.is_active ? 'Active' : 'Revoked'}
                          </span>
                        </div>
                        <div className="flex flex-col gap-1 mt-1">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-[#888c99] w-24">Secret:</span>
                            <span className="text-sm text-red-300/80 font-mono">
                              sk_{selectedApp?.is_live ? 'live' : 'test'}_••••••••{key.secret_hint}
                            </span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-[#888c99] w-24">Publishable:</span>
                            <span className="text-sm text-emerald-300/80 font-mono">
                              pk_{selectedApp?.is_live ? 'live' : 'test'}_••••••••{key.publishable_hint}
                            </span>
                          </div>
                        </div>
                        {key.last_used_at && (
                          <span className="text-xs text-[#888c99] block mt-2">
                            Last used: {new Date(key.last_used_at).toLocaleString()}
                          </span>
                        )}
                      </div>
                    </div>
                    {key.is_active && (
                      <button
                        onClick={() => handleRevokeKey(key.id)}
                        className="px-3 py-1.5 text-xs font-semibold text-red-400 hover:bg-red-500/10 rounded-lg transition-colors border border-transparent hover:border-red-500/20"
                      >
                        Revoke Pair
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Logs Tab */}
        {detailTab === 'logs' && (
          <div className="flex flex-col gap-4">
            {logs.length === 0 ? (
              <div className="p-8 bg-[#1c1f26] border border-[#22252e] rounded-2xl text-center text-[#888c99]">
                No API requests logged yet.
              </div>
            ) : (
              <>
                <div className="w-full overflow-x-auto rounded-xl border border-[#22252e]">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="bg-[#1c1f26] border-b border-[#22252e]">
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Method</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Path</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Status</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider text-right">Latency</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider text-right">Time</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-[#22252e] bg-[#13151a]">
                      {logs.map(log => (
                        <tr key={log.id} className="hover:bg-[#1c1f26]/60 transition-colors">
                          <td className="py-3 px-4">
                            <span className={`font-mono text-xs font-bold px-2 py-0.5 rounded ${log.method === 'GET' ? 'bg-blue-500/10 text-blue-400' : 'bg-amber-500/10 text-amber-400'
                              }`}>
                              {log.method}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-sm text-white font-mono">{log.path}</td>
                          <td className="py-3 px-4">
                            <span className={`text-sm font-bold ${getStatusColor(log.status_code)}`}>
                              {log.status_code}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-sm text-[#888c99] text-right">{log.latency_ms}ms</td>
                          <td className="py-3 px-4 text-sm text-[#888c99] text-right">
                            {new Date(log.created_at).toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {logsTotal > 50 && (
                  <div className="flex items-center justify-center gap-2">
                    <button
                      onClick={() => loadLogs(selectedApp.id, logsPage - 1)}
                      disabled={logsPage <= 1}
                      className="px-4 py-2 text-sm bg-[#1c1f26] border border-[#22252e] rounded-xl disabled:opacity-50 text-white hover:bg-[#22252e] transition-colors"
                    >
                      Previous
                    </button>
                    <span className="text-sm text-[#888c99]">Page {logsPage}</span>
                    <button
                      onClick={() => loadLogs(selectedApp.id, logsPage + 1)}
                      disabled={logsPage * 50 >= logsTotal}
                      className="px-4 py-2 text-sm bg-[#1c1f26] border border-[#22252e] rounded-xl disabled:opacity-50 text-white hover:bg-[#22252e] transition-colors"
                    >
                      Next
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        )}

        {/* Webhooks Tab */}
        {detailTab === 'webhooks' && (
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-[#888c99]">Receive real-time notifications for events.</p>
              <button
                onClick={() => setShowWebhookForm(!showWebhookForm)}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-xl transition-colors"
              >
                <Icon icon="hugeicons:add-circle-half-dot" className="size-4" />
                Add Webhook
              </button>
            </div>

            {webhooks.length === 0 ? (
              <div className="p-8 bg-[#1c1f26] border border-[#22252e] rounded-2xl text-center text-[#888c99]">
                No webhooks registered yet.
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {webhooks.map(wh => (
                  <div key={wh.id} className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="size-10 rounded-xl bg-purple-500/10 flex items-center justify-center">
                        <Icon icon="hugeicons:notification-03" className="size-5 text-purple-400" />
                      </div>
                      <div className="flex flex-col gap-1">
                        <span className="text-sm text-white font-mono truncate max-w-[300px]">{wh.url}</span>
                        <div className="flex flex-wrap gap-1">
                          {wh.events.map(e => (
                            <span key={e} className="px-1.5 py-0.5 bg-[#13151a] text-[#888c99] text-[10px] rounded border border-[#22252e]">{e}</span>
                          ))}
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() => handleDeleteWebhook(wh.id)}
                      className="size-8 rounded-lg hover:bg-red-500/10 flex items-center justify-center transition-colors"
                    >
                      <Icon icon="hugeicons:delete-02" className="size-4 text-red-400" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="flex flex-col w-full max-w-[1400px] mx-auto animate-in fade-in duration-500 pb-20 font-sans gap-8">
      {renderKeysModal()}
      {renderWebhookSecretModal()}
      {renderCreateAppModal()}
      {renderWebhookModal()}
      {renderRegenerateModal()}

      {/* Page Header */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] p-8 shadow-2xl relative overflow-hidden">
        <div className="relative z-10 flex items-center gap-4">
          <div className="size-14 rounded-2xl bg-gradient-to-br from-blue-600  flex items-center justify-center ">
            <Icon icon="hugeicons:api" className="size-7 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white tracking-tight">Developer Portal</h1>
            <p className="text-sm text-[#888c99]">Build integrations with our crypto exchange API.</p>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] p-6 sm:p-8 shadow-xl">
        {activeTab === 'apps' && renderApps()}
        {activeTab === 'detail' && renderDetail()}
      </div>
    </div>
  );
}
