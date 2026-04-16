/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import {
  Bell,
  Send,
  Trash2,
  Calendar,
  Mail,
  MessageSquare,
  Clock,
  AlertCircle,
  Plus,
  CheckCircle,
  XCircle
} from "lucide-react";

interface Notification {
  id: string;
  destination: string;
  channel: "email" | "telegram";
  message: string;
  data_sent_at: string;
  status: "pending" | "sent" | "canceled" | "created" | "failed";
  created_at: string;
}

export default function App() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    destination: "",
    channel: "email" as const,
    message: "",
    sendAt: ""
  });

  useEffect(() => {
    const now = new Date();
    now.setSeconds(0, 0);
    const tzOffset = now.getTimezoneOffset() * 60000;
    const localISOTime = new Date(now.getTime() - tzOffset).toISOString().slice(0, 16);
    setFormData(prev => ({ ...prev, sendAt: localISOTime }));

    fetchNotifications();
    const interval = setInterval(fetchNotifications, 5000);
    return () => clearInterval(interval);
  }, []);

    const fetchNotifications = async () => {
        try {
            const res = await fetch("/all", {
                headers: { Accept: "application/json" }
            });

            if (!res.ok) {
                console.error("Failed to fetch notifications:", res.status);
                return;
            }

            const contentType = res.headers.get("content-type") || "";
            if (!contentType.includes("application/json")) {
                const text = await res.text();
                console.error("Expected JSON but received:", text.slice(0, 100));
                return;
            }

            const data = await res.json();
            const list = Array.isArray(data) ? data : (data.notes || data.items || data.data || []);
            setNotifications(list);
        } catch (error) {
            console.error("Failed to fetch notifications:", error);
        } finally {
            setLoading(false);
        }
    };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await fetch("/create", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Accept": "application/json"
        },
        body: JSON.stringify({
          destination: formData.destination,
          channel: formData.channel,
          message: formData.message,
          data_sent_at: new Date(formData.sendAt).toISOString()
        })
      });
      if (res.ok) {
        setFormData({
          ...formData,
          destination: "",
          message: ""
        });
        await fetchNotifications();
      }
    } catch (error) {
      console.error("Failed to create notification:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = async (id: string) => {
    try {
      const res = await fetch(`/cancel/${id}`, { method: "DELETE" });
      if (res.ok) {
        await fetchNotifications();
      }
    } catch (error) {
      console.error("Failed to cancel notification:", error);
    }
  };

    const pendingCount = notifications.filter(n => n.status === "pending" || n.status === "created").length;

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      {/* Navbar */}
      <nav className="h-16 bg-surface border-b border-border flex items-center justify-between px-10 flex-none">
        <div className="flex items-center gap-3">
          <div className="w-6 h-6 bg-primary rounded-md flex items-center justify-center">
            <Bell size={14} className="text-white" />
          </div>
          <span className="font-bold text-lg tracking-tight">Delayed Notifier</span>
        </div>
        <div className="flex items-center gap-5 text-sm text-text-sub font-medium">
          <span>Project Alpha</span>
          <div className="w-8 h-8 rounded-full bg-gray-200 border border-border overflow-hidden">
            <img
              src="https://picsum.photos/seed/user/32/32"
              alt="User"
              className="w-full h-full object-cover"
              referrerPolicy="no-referrer"
            />
          </div>
        </div>
      </nav>

      {/* Main Container */}
      <div className="flex-1 grid grid-cols-[380px_1fr] gap-8 p-10 overflow-hidden">
        {/* Sidebar - Form */}
        <aside className="minimal-card p-8 self-start">
          <h2 className="text-lg font-semibold mb-6">Schedule Notification</h2>

          <form onSubmit={handleCreate} className="space-y-5">
            <div>
              <label className="label-minimal">Destination</label>
              <input
                type="text"
                required
                className="input-minimal"
                placeholder="Email or username"
                value={formData.destination}
                onChange={e => setFormData({ ...formData, destination: e.target.value })}
              />
            </div>

            <div>
              <label className="label-minimal">Channel</label>
              <select
                className="input-minimal appearance-none"
                value={formData.channel}
                onChange={e => setFormData({ ...formData, channel: e.target.value as any })}
              >
                <option value="email">Email</option>
                <option value="telegram">Telegram</option>
              </select>
            </div>

            <div>
              <label className="label-minimal">Scheduled Date & Time</label>
              <input
                type="datetime-local"
                required
                className="input-minimal"
                value={formData.sendAt}
                onChange={e => setFormData({ ...formData, sendAt: e.target.value })}
              />
            </div>

            <div>
              <label className="label-minimal">Message Body</label>
              <textarea
                required
                rows={4}
                className="input-minimal resize-none"
                placeholder="Type your message here..."
                value={formData.message}
                onChange={e => setFormData({ ...formData, message: e.target.value })}
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full bg-primary text-white rounded-lg py-3 font-semibold text-sm hover:opacity-90 transition-opacity flex items-center justify-center gap-2"
            >
              {isSubmitting ? (
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                "Create Schedule"
              )}
            </button>
          </form>
        </aside>

        {/* Queue Container */}
        <main className="flex flex-col gap-4 overflow-hidden">
          <div className="flex justify-between items-center mb-2 flex-none">
            <h2 className="text-lg font-semibold">Active Queue</h2>
            <span className="text-sm bg-border px-2 py-0.5 rounded-full font-medium">
              {pendingCount} Pending
            </span>
          </div>

          <div className="flex-1 overflow-y-auto pr-2 space-y-4 pb-10 custom-scrollbar">
            {loading ? (
              <div className="flex items-center justify-center h-40 text-text-sub font-medium">
                Loading queue...
              </div>
            ) : notifications.length === 0 ? (
              <div className="minimal-card border-dashed p-12 text-center flex flex-col items-center justify-center">
                <AlertCircle className="text-gray-300 mb-3" size={32} />
                <p className="text-text-sub font-medium">No active schedules in the queue.</p>
              </div>
            ) : (
              <AnimatePresence mode="popLayout">
                {notifications.slice().sort((a,b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((note) => (
                  <motion.div
                    layout
                    key={note.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.98 }}
                    className="minimal-card p-5 relative flex flex-col gap-3 group"
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex items-center gap-3">
                        <span className={`badge-minimal ${note.channel === 'email' ? 'bg-[#FEF3C7] text-[#D97706]' : 'bg-[#E0F2FE] text-[#0284C7]'}`}>
                          {note.channel === 'email' ? 'Email' : 'Telegram'}
                        </span>
                        <span className={`badge-minimal ${
                          (note.status === 'pending' || note.status === 'created') ? 'status-scheduled' : 
                          note.status === 'sent' ? 'status-sent' : 
                          note.status === 'failed' ? 'bg-red-100 text-red-700' : 'status-canceled'
                        }`}>
                          { (note.status === 'pending' || note.status === 'created') ? 'Scheduled' : 
                            note.status === 'sent' ? 'Sent' : 
                            note.status === 'failed' ? 'Failed' : 'Canceled'}
                        </span>
                      </div>

                      {(note.status === 'pending' || note.status === 'created') && (
                        <button
                          onClick={() => handleCancel(note.id)}
                          className="px-3 py-1.5 border border-border rounded-md text-xs font-medium text-red-500 hover:bg-red-50 transition-colors"
                        >
                          Cancel
                        </button>
                      )}
                    </div>

                    <div className="text-[15px] leading-relaxed text-text-main pr-16 line-clamp-3">
                      {note.message}
                    </div>

                    <div className="flex items-center gap-6 text-[12px] text-text-sub font-medium">
                      <div className="flex items-center gap-1.5">
                        <span className="text-text-sub/60 uppercase text-[10px] tracking-wider">To:</span>
                        {note.destination}
                      </div>
                      <div className="flex items-center gap-1.5 ml-auto">
                        <Calendar size={12} className="opacity-50" />
                        {new Date(note.data_sent_at).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' })}
                      </div>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}


