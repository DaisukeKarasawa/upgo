import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getChange, syncChange } from "../services/api";
import StatusBadge from "../components/StatusBadge";
import ManualSyncButton from "../components/ManualSyncButton";
import Markdown from "../components/Markdown";

interface File {
  id: number;
  file_path: string;
  status?: string;
  lines_inserted: number;
  lines_deleted: number;
  size_delta: number;
}

interface Revision {
  id: number;
  revision_id: string;
  patchset_num: number;
  uploader_name?: string;
  uploader_email?: string;
  created: string;
  commit_message?: string;
  files?: File[];
}

interface Comment {
  id: number;
  comment_id: string;
  file_path?: string;
  line?: number;
  author_name: string;
  author_email?: string;
  message: string;
  created: string;
  updated: string;
  in_reply_to?: string;
  unresolved: boolean;
}

interface Label {
  id: number;
  label_name: string;
  value: number;
  account_name?: string;
  account_email?: string;
  granted_on: string;
}

interface Message {
  id: number;
  message_id: string;
  author_name?: string;
  author_email?: string;
  message: string;
  date: string;
  revision_number?: number;
}

interface ChangeData {
  id: number;
  change_id: string;
  change_number: number;
  project: string;
  branch: string;
  status: string;
  subject: string;
  message?: string;
  owner_name: string;
  owner_email?: string;
  created: string;
  updated: string;
  submitted?: string;
  last_synced_at?: string;
  revisions?: Revision[];
  comments?: Comment[];
  labels?: Label[];
  messages?: Message[];
}

export default function Detail() {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<ChangeData | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<
    "revisions" | "files" | "comments" | "labels" | "messages"
  >("revisions");

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const result = await getChange(parseInt(id));
      setData(result as ChangeData);
    } catch (error) {
      console.error("Failed to fetch data", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    if (!data) return;
    try {
      await syncChange(data.change_number);
      setTimeout(() => {
        loadData();
      }, 2000);
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
  const hasComments = data.comments && data.comments.length > 0;
  const hasLabels = data.labels && data.labels.length > 0;
  const hasMessages = data.messages && data.messages.length > 0;

  const gerritUrl = `https://go-review.googlesource.com/c/go/+/${data.change_number}`;

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
              <span className="text-gray-400 mr-2">#{data.change_number}</span>
              {data.subject}
            </h1>
            <div className="flex-shrink-0">
              <ManualSyncButton onSync={handleSync} hasUpdates={false} />
            </div>
          </div>

          {data.message && (
            <div className="mb-8">
              <Markdown>{data.message}</Markdown>
            </div>
          )}

          <div className="border-t border-gray-100 pt-6">
            <div className="text-sm text-gray-400 space-y-1 font-light">
              <p>Owner: {data.owner_name}</p>
              <p>Branch: {data.branch}</p>
              <p>
                Created: {new Date(data.created).toLocaleString("en-US")}
              </p>
              <p>
                Updated: {new Date(data.updated).toLocaleString("en-US")}
              </p>
              {data.submitted && (
                <p>
                  Submitted: {new Date(data.submitted).toLocaleString("en-US")}
                </p>
              )}
              <a
                href={gerritUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-400 hover:text-gray-900 transition-all duration-300 ease-out inline-block mt-3 group"
              >
                Open on Gerrit{" "}
                <span className="inline-block transition-transform duration-300 ease-out group-hover:translate-x-1">
                  →
                </span>
              </a>
            </div>
          </div>
        </div>

        <div>
          <nav className="flex gap-8 mb-8 border-b border-gray-100">
            <button
              onClick={() => setActiveTab("revisions")}
              className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                activeTab === "revisions"
                  ? "text-gray-900"
                  : "text-gray-400 hover:text-gray-600"
              }`}
            >
              Patchsets ({data.revisions?.length || 0})
              {activeTab === "revisions" && (
                <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
              )}
            </button>
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
            {activeTab === "revisions" && (
              <div className="space-y-8">
                {hasRevisions ? (
                  data.revisions?.map((revision) => (
                    <div key={revision.id} className="border border-gray-100 rounded-lg p-6">
                      <div className="flex items-center justify-between mb-4">
                        <h4 className="text-base font-normal text-gray-900">
                          Patchset {revision.patchset_num}
                        </h4>
                        <span className="text-xs text-gray-400 font-light">
                          {new Date(revision.created).toLocaleString("en-US")}
                        </span>
                      </div>
                      {revision.uploader_name && (
                        <p className="text-sm text-gray-500 mb-3">
                          Uploaded by: {revision.uploader_name}
                        </p>
                      )}
                      {revision.commit_message && (
                        <div className="bg-gray-50 p-4 rounded mb-4">
                          <pre className="text-sm whitespace-pre-wrap text-gray-700 font-mono">
                            {revision.commit_message}
                          </pre>
                        </div>
                      )}
                      {revision.files && revision.files.length > 0 && (
                        <div>
                          <h5 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-3">
                            Files ({revision.files.length})
                          </h5>
                          <div className="space-y-2">
                            {revision.files.map((file) => (
                              <div key={file.id} className="flex items-center justify-between text-sm py-2 border-b border-gray-50 last:border-0">
                                <span className="text-gray-700 font-mono text-xs">{file.file_path}</span>
                                <div className="flex items-center gap-3 text-xs">
                                  <span className="text-green-600">+{file.lines_inserted}</span>
                                  <span className="text-red-600">-{file.lines_deleted}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No patchsets available. Click "Sync" to fetch details.
                  </div>
                )}
              </div>
            )}

            {activeTab === "comments" && (
              <div className="space-y-6">
                {hasComments ? (
                  data.comments?.map((comment) => (
                    <div
                      key={comment.id}
                      className={`p-4 rounded-lg ${comment.unresolved ? 'bg-yellow-50 border border-yellow-100' : 'bg-gray-50'}`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-gray-900 font-light">
                          {comment.author_name}
                        </span>
                        <div className="flex items-center gap-2">
                          {comment.unresolved && (
                            <span className="text-xs text-yellow-600 font-light">Unresolved</span>
                          )}
                          <span className="text-xs text-gray-400 font-light">
                            {new Date(comment.created).toLocaleString("en-US")}
                          </span>
                        </div>
                      </div>
                      {comment.file_path && (
                        <p className="text-xs text-gray-500 mb-2 font-mono">
                          {comment.file_path}{comment.line ? `:${comment.line}` : ''}
                        </p>
                      )}
                      <Markdown className="text-sm">{comment.message}</Markdown>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    No comments.
                  </div>
                )}
              </div>
            )}

            {activeTab === "labels" && (
              <div className="space-y-4">
                {hasLabels ? (
                  data.labels?.map((label) => (
                    <div key={label.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                      <div>
                        <span className="text-sm text-gray-900 font-light">{label.label_name}</span>
                        <span className={`ml-2 text-sm font-medium ${label.value > 0 ? 'text-green-600' : label.value < 0 ? 'text-red-600' : 'text-gray-500'}`}>
                          {label.value > 0 ? `+${label.value}` : label.value}
                        </span>
                      </div>
                      <div className="text-right">
                        {label.account_name && (
                          <p className="text-sm text-gray-500">{label.account_name}</p>
                        )}
                        <p className="text-xs text-gray-400">
                          {new Date(label.granted_on).toLocaleString("en-US")}
                        </p>
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

            {activeTab === "messages" && (
              <div className="space-y-6">
                {hasMessages ? (
                  data.messages?.map((msg) => (
                    <div key={msg.id} className="p-4 bg-gray-50 rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-gray-900 font-light">
                          {msg.author_name || 'System'}
                        </span>
                        <div className="flex items-center gap-2">
                          {msg.revision_number && (
                            <span className="text-xs text-gray-500">PS{msg.revision_number}</span>
                          )}
                          <span className="text-xs text-gray-400 font-light">
                            {new Date(msg.date).toLocaleString("en-US")}
                          </span>
                        </div>
                      </div>
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
