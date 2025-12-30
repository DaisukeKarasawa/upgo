import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getPR } from "../services/api";
import StatusBadge from "../components/StatusBadge";

interface Summary {
  description_summary?: string;
  diff_summary?: string;
  diff_explanation?: string;
  comments_summary?: string;
  discussion_summary?: string;
  merge_reason?: string;
  close_reason?: string;
}

interface Diff {
  file_path: string;
  diff_text: string;
}

interface Comment {
  github_id: number;
  body: string;
  author: string;
  created_at: string;
  updated_at: string;
}

interface DetailData {
  id: number;
  title: string;
  body: string;
  state: string;
  author: string;
  created_at: string;
  updated_at: string;
  url: string;
  summary?: Summary;
  diffs?: Diff[];
  comments?: Comment[];
}

export default function Detail() {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<DetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<
    "summary" | "diffs" | "comments" | "analysis"
  >("summary");

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const result = await getPR(parseInt(id));
      setData(result as DetailData);
    } catch (error) {
      console.error("データの取得に失敗しました", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-400 text-sm font-light">読み込み中...</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="text-gray-400 text-sm font-light">
          データが見つかりません
        </div>
      </div>
    );
  }

  const hasSummary = data.summary && Object.keys(data.summary).length > 0;
  const hasDiffs = data.diffs && data.diffs.length > 0;
  const hasComments = data.comments && data.comments.length > 0;

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
          ダッシュボードに戻る
        </Link>

        <div className="mb-12">
          <div className="flex items-center gap-3 mb-6">
            <StatusBadge state={data.state} alwaysColored={true} />
            <h1 className="text-3xl font-light text-gray-900 tracking-tight">
              {data.title}
            </h1>
          </div>

          <div className="prose max-w-none mb-8">
            <p className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
              {data.body}
            </p>
          </div>

          <div className="border-t border-gray-100 pt-6">
            <div className="text-sm text-gray-400 space-y-1 font-light">
              <p>作成者: {data.author}</p>
              <p>
                作成日時: {new Date(data.created_at).toLocaleString("ja-JP")}
              </p>
              <p>
                更新日時: {new Date(data.updated_at).toLocaleString("ja-JP")}
              </p>
              {data.url && (
                <a
                  href={data.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-gray-400 hover:text-gray-900 transition-all duration-300 ease-out inline-block mt-3 group"
                >
                  GitHubで開く{" "}
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
            {hasSummary && (
              <button
                onClick={() => setActiveTab("summary")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "summary"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                要約
                {activeTab === "summary" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
            {hasDiffs && (
              <button
                onClick={() => setActiveTab("diffs")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "diffs"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                Diff
                {activeTab === "diffs" && (
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
            {(data.state === "merged" || data.state === "closed") && (
              <button
                onClick={() => setActiveTab("analysis")}
                className={`pb-3 text-sm font-light transition-all duration-300 ease-out relative ${
                  activeTab === "analysis"
                    ? "text-gray-900"
                    : "text-gray-400 hover:text-gray-600"
                }`}
              >
                分析結果
                {activeTab === "analysis" && (
                  <span className="absolute bottom-0 left-0 right-0 h-px bg-gray-900 animate-[slideIn_0.3s_ease-out]" />
                )}
              </button>
            )}
          </nav>

          <div>
            {/* 要約タブ */}
            {activeTab === "summary" && hasSummary && (
              <div className="space-y-10">
                {data.summary?.description_summary && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      説明の要約
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.description_summary}
                    </div>
                  </div>
                )}

                {data.summary?.diff_summary && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      差分の要約
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.diff_summary}
                    </div>
                  </div>
                )}

                {data.summary?.diff_explanation && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      差分の解説
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.diff_explanation}
                    </div>
                  </div>
                )}

                {data.summary?.comments_summary && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      コメントの要約
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.comments_summary}
                    </div>
                  </div>
                )}

                {data.summary?.discussion_summary && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      議論の要約
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.discussion_summary}
                    </div>
                  </div>
                )}

                {!hasSummary && (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    要約データがまだ生成されていません。しばらくお待ちください。
                  </div>
                )}
              </div>
            )}

            {/* 差分タブ */}
            {activeTab === "diffs" && (
              <div className="space-y-8">
                {hasDiffs ? (
                  data.diffs?.map((diff, index) => (
                    <div key={index}>
                      <h4 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-3">
                        {diff.file_path}
                      </h4>
                      <div className="bg-gray-50 p-4 rounded">
                        <pre className="text-xs overflow-x-auto font-mono">
                          <code className="whitespace-pre-wrap text-gray-700">
                            {diff.diff_text}
                          </code>
                        </pre>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    差分データがありません。
                  </div>
                )}
              </div>
            )}

            {/* コメントタブ */}
            {activeTab === "comments" && (
              <div className="space-y-8">
                {hasComments ? (
                  data.comments?.map((comment) => (
                    <div
                      key={comment.github_id}
                      className="pb-8 border-b border-gray-100 last:border-0 group hover:pl-2 transition-all duration-300 ease-out"
                    >
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm text-gray-900 font-light group-hover:text-gray-700 transition-colors duration-300">
                          {comment.author}
                        </span>
                        <span className="text-xs text-gray-400 font-light">
                          {new Date(comment.created_at).toLocaleString("ja-JP")}
                        </span>
                      </div>
                      <div className="prose max-w-none">
                        <p className="text-gray-700 whitespace-pre-wrap leading-relaxed text-sm font-light group-hover:text-gray-800 transition-colors duration-300">
                          {comment.body}
                        </p>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    コメントがありません。
                  </div>
                )}
              </div>
            )}

            {/* 分析結果タブ */}
            {activeTab === "analysis" && (
              <div className="space-y-10">
                {data.state === "merged" && data.summary?.merge_reason && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      Merge理由の分析
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.merge_reason}
                    </div>
                  </div>
                )}

                {data.state === "closed" && data.summary?.close_reason && (
                  <div>
                    <h3 className="text-sm text-gray-400 font-light uppercase tracking-wider mb-4">
                      Close理由の分析
                    </h3>
                    <div className="text-gray-700 whitespace-pre-wrap leading-relaxed font-light">
                      {data.summary.close_reason}
                    </div>
                  </div>
                )}

                {!data.summary?.merge_reason && !data.summary?.close_reason && (
                  <div className="text-center py-20 text-gray-400 text-sm font-light">
                    分析結果がまだ生成されていません。
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
