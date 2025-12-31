import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getChange, syncChange, getChangeUpdateStatus } from "../services/api";
import StatusBadge from "../components/StatusBadge";
import ManualSyncButton from "../components/ManualSyncButton";
import Markdown from "../components/Markdown";

interface Revision {
  id: number;
  patch_set_number: number;
  revision_sha: string;
  uploader: string;
  created_at: string;
  kind?: string;
  commit_message?: string;
}

interface File {
  file_path: string;
  status?: string;
  old_path?: string;
  lines_inserted: number;
  lines_deleted: number;
  size_delta: number;
  size: number;
  binary: boolean;
}

interface Comment {
  comment_id: string;
  file_path?: string;
  line?: number;
  patch_set_number: number;
  message: string;
  author: string;
  created_at: string;
  updated_at: string;
  in_reply_to?: string;
  unresolved: boolean;
}

interface Label {
  label_name: string;
  account: string;
  value: number;
  date: string;
}

interface Message {
  message_id: string;
  author: string;
  message: string;
  date: string;
  revision_number?: number;
}

interface DetailData {
  id: number;
  change_number: number;
  change_id: string;
  project: string;
  branch: string;
  subject: string;
  message?: string;
  status: string;
  owner: string;
  created_at: string;
  updated_at: string;
  submitted_at?: string;
  url: string;
  revisions?: Revision[];
  files?: File[];
  comments?: Comment[];
  labels?: Label[];
  messages?: Message[];
}

export default function Detail() {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<DetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [hasUpdates, setHasUpdates] = useState(false);
  const [activeTab, setActiveTab] = useState<
    "overview" | "revisions" | "files" | "comments" | "labels" | "messages"
  >("overview");

  useEffect(() => {
    loadData();
    if (id) {
      checkUpdates();

      // Poll for updates every 30 seconds
      const interval = setInterval(() => {
        checkUpdates();
      }, 30000);

      return () => clearInterval(interval);
    }
  }, [id]);

  const checkUpdates = async () => {
    if (!id) return;
    try {
      const status = await getChangeUpdateStatus(parseInt(id));
      setHasUpdates(status.updated_since_last_sync || false);
    } catch (error) {
      console.error("Failed to check for updates", error);
    }
  };

  const loadData = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const result = await getChange(parseInt(id));
      setData(result as DetailData);
    } catch (error) {
      console.error("Failed to fetch data", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    if (!id) return;
    try {
      await syncChange(parseInt(id));
      setTimeout(() => {
        loadData();
        checkUpdates(); // Check for updates after sync
      }, 2000); // Reload after 2 seconds
    } catch (error) {
      console.error("Failed to sync", error);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-400 text-sm font-light">Loading...</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-400 text-sm font-light">Data not found</div>
      </div>
    );
  }

  const hasRevisions = data.revisions && data.revisions.length > 0;
  const hasFiles = data.files && data.files.length > 0;
  const hasComments = data.comments && data.comments.length > 0;
  const hasLabels = data.labels && data.labels.length > 0;
  const hasMessages = data.messages && data.messages.length > 0;

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-5xl mx-auto px-6 py-12">
        <Link
          to="/"
          className="text-gray-400 hover:text-gray-900 transition-all duration-300 ease-out mb-8 inline-block text-sm font-light group"
        >
          <span className="inline-block transition-transform duration-300 ease-out group-hover:-translate-x-1">
            ←
          </span>{" "}
          Back to Dashboard
        </Link>

        <div className="mb-12">
          <div className="flex items-center gap-3 mb-6">
            <StatusBadge state={data.status} alwaysColored={true} />
            <h1 className="text-3xl font-light text-gray-900 tracking-tight flex-1">
              {data.subject}
            </h1>
            <div className="flex-shrink-0">
              <ManualSyncButton onSync={handleSync} hasUpdates={hasUpdates} />
            </div>
          </div>

          {data.message && (
            <div className="mb-8">
              <Markdown>{data.message}</Markdown>
            </div>
          )}

          <div className="border-t border-gray-100 pt-6">
            <div className="text-sm text-gray-400 space-y-1 font-light">
              <p>Change Number: {data.change_number}</p>
              <p>Change ID: {data.change_id}</p>
              <p>Project: {data.project}</p>
              <p>Branch: {data.branch}</p>
              <p>Owner: {data.owner}</p>
              <p>
                Created: {new Date(data.created_at).toLocaleString("en-US")}
              </p>
              <p>
                Updated: {new Date(data.updated_at).toLocaleString("en-US")}
              </p>
              {data.submitted_at && (
                <p>
                  Submitted: {new Date(data.submitted_at).toLocaleString("en-US")}
                </p>
              )}
              {data.url && (
                <a
                  href={data.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-gray-400 hover:text-gray-900 transition-all duration-300 ease-out inline-block mt-3 group"
                >
                  Open on Gerrit{" "}
                  <span className="inline-block transition-transform duration-300 ease-out group-hover:translate-x-1">
                    →
                  </span>
                </a>
              )}
            </div>
          </div>
        </div>

        {/* タブナビゲーション */}
        <div>
          <nav className="flex gap-8 mb-8 border-b border-gray-100">
            <button
              onClick={() => setActiveTab("overview")}
              className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                activeTab === "overview"
                  ? "text-gray-900"
                  : "text-gray-400 hover:text-gray-600"
              }`}
            >
              Overview
              {activeTab === "overview" && (
                <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
              )}
            </button>
            {hasRevisions && (
              <button
                onClick={() => setActiveTab("revisions")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "revisions"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Revisions ({data.revisions?.length || 0})
                {activeTab === "revisions" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
            {hasFiles && (
              <button
                onClick={() => setActiveTab("files")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "files"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Files ({data.files?.length || 0})
                {activeTab === "files" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
            {hasComments && (
              <button
                onClick={() => setActiveTab("comments")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "comments"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Comments ({data.comments?.length || 0})
                {activeTab === "comments" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
            {hasLabels && (
              <button
                onClick={() => setActiveTab("labels")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "labels"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Labels ({data.labels?.length || 0})
                {activeTab === "labels" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
            {hasMessages && (
              <button
                onClick={() => setActiveTab("messages")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "messages"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Messages ({data.messages?.length || 0})
                {activeTab === "messages" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
          </nav>

          <div>
            {/* Overview Tab */}
            {activeTab === "overview" && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                    Change Information
                  </h3>
                  <div className="text-sm text-gray-700 space-y-2">
                    <p><strong>Change Number:</strong> {data.change_number}</p>
                    <p><strong>Change ID:</strong> {data.change_id}</p>
                    <p><strong>Project:</strong> {data.project}</p>
                    <p><strong>Branch:</strong> {data.branch}</p>
                    <p><strong>Status:</strong> {data.status}</p>
                    <p><strong>Owner:</strong> {data.owner}</p>
                  </div>
                </div>
                {hasLabels && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      Labels Summary
                    </h3>
                    <div className="space-y-2">
                      {data.labels?.slice(0, 5).map((label, idx) => (
                        <div key={idx} className="text-sm text-gray-700">
                          <span className="font-medium">{label.label_name}:</span> {label.value > 0 ? '+' : ''}{label.value} by {label.account}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Revisions Tab */}
            {activeTab === "revisions" && (
              <div className="space-y-6">
                {hasRevisions ? (
                  data.revisions?.map((rev) => (
                    <div key={rev.id} className="border-b border-gray-100 pb-6 last:border-0">
                      <div className="flex items-center justify-between mb-3">
                        <div>
                          <h4 className="text-base font-medium text-gray-900">
                            Patch Set {rev.patch_set_number}
                          </h4>
                          <p className="text-sm text-gray-500 font-mono">
                            {rev.revision_sha.substring(0, 12)}
                          </p>
                        </div>
                        <div className="text-sm text-gray-400">
                          {new Date(rev.created_at).toLocaleString("en-US")}
                        </div>
                      </div>
                      <div className="text-sm text-gray-600 space-y-1">
                        <p><strong>Uploader:</strong> {rev.uploader}</p>
                        {rev.kind && <p><strong>Kind:</strong> {rev.kind}</p>}
                        {rev.commit_message && (
                          <div className="mt-3">
                            <p className="font-medium mb-1">Commit Message:</p>
                            <Markdown className="text-sm">{rev.commit_message}</Markdown>
                          </div>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No revisions available.
                  </div>
                )}
              </div>
            )}

            {/* Files Tab */}
            {activeTab === "files" && (
              <div className="space-y-4">
                {hasFiles ? (
                  data.files?.map((file, idx) => (
                    <div key={idx} className="border-b border-gray-100 pb-4 last:border-0">
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="text-sm font-medium text-gray-900">
                          {file.file_path}
                        </h4>
                        {file.status && (
                          <span className="text-xs text-gray-400 uppercase">
                            {file.status}
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-gray-500 space-x-4">
                        <span>+{file.lines_inserted}</span>
                        <span>-{file.lines_deleted}</span>
                        {file.binary && <span className="text-orange-500">Binary</span>}
                      </div>
                      {file.old_path && file.old_path !== file.file_path && (
                        <div className="text-xs text-gray-400 mt-1">
                          Renamed from: {file.old_path}
                        </div>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No files available.
                  </div>
                )}
              </div>
            )}

            {/* Comments Tab */}
            {activeTab === "comments" && (
              <div className="space-y-8">
                {hasComments ? (
                  data.comments?.map((comment) => (
                    <div
                      key={comment.comment_id}
                      className="pb-8 border-b border-gray-100 last:border-0 group transition-all duration-300 ease-out relative rounded-lg overflow-hidden"
                    >
                      <div className="p-4 group-hover:px-6 group-hover:py-5 transition-all duration-300 ease-out">
                        <div className="flex items-center justify-between mb-3">
                          <div className="flex items-center gap-3">
                            <span className="text-sm text-gray-900 font-light group-hover:text-gray-700 transition-colors duration-300">
                              {comment.author}
                            </span>
                            {comment.file_path && (
                              <span className="text-xs text-gray-400 font-mono">
                                {comment.file_path}
                                {comment.line && `:${comment.line}`}
                              </span>
                            )}
                            {comment.unresolved && (
                              <span className="text-xs text-orange-500">Unresolved</span>
                            )}
                          </div>
                          <span className="text-xs text-gray-400 font-light group-hover:text-gray-600 transition-colors duration-300">
                            {new Date(comment.created_at).toLocaleString("en-US")}
                          </span>
                        </div>
                        <div className="text-sm mb-2">
                          <span className="text-gray-400">Patch Set {comment.patch_set_number}</span>
                        </div>
                        <Markdown className="text-sm">{comment.message}</Markdown>
                        {comment.in_reply_to && (
                          <div className="text-xs text-gray-400 mt-2">
                            In reply to: {comment.in_reply_to}
                          </div>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No comments.
                  </div>
                )}
              </div>
            )}

            {/* Labels Tab */}
            {activeTab === "labels" && (
              <div className="space-y-4">
                {hasLabels ? (
                  data.labels?.map((label, idx) => (
                    <div key={idx} className="border-b border-gray-100 pb-4 last:border-0">
                      <div className="flex items-center justify-between">
                        <div>
                          <span className="text-sm font-medium text-gray-900">
                            {label.label_name}
                          </span>
                          <span className={`ml-3 text-sm ${label.value > 0 ? 'text-green-600' : label.value < 0 ? 'text-red-600' : 'text-gray-400'}`}>
                            {label.value > 0 ? '+' : ''}{label.value}
                          </span>
                        </div>
                        <div className="text-xs text-gray-400">
                          {label.account} • {new Date(label.date).toLocaleString("en-US")}
                        </div>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No labels.
                  </div>
                )}
              </div>
            )}

            {/* Messages Tab */}
            {activeTab === "messages" && (
              <div className="space-y-6">
                {hasMessages ? (
                  data.messages?.map((msg) => (
                    <div key={msg.message_id} className="border-b border-gray-100 pb-6 last:border-0">
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm text-gray-900 font-light">
                          {msg.author}
                        </span>
                        <span className="text-xs text-gray-400 font-light">
                          {new Date(msg.date).toLocaleString("en-US")}
                        </span>
                      </div>
                      {msg.revision_number && (
                        <div className="text-xs text-gray-400 mb-2">
                          Patch Set {msg.revision_number}
                        </div>
                      )}
                      <Markdown className="text-sm">{msg.message}</Markdown>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No messages.
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
