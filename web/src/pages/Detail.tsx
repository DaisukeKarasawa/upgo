import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getPR, getIssue } from '../services/api'
import StatusBadge from '../components/StatusBadge'

interface DetailProps {
  type: 'pr' | 'issue'
}

interface Summary {
  description_summary?: string
  diff_summary?: string
  diff_explanation?: string
  comments_summary?: string
  discussion_summary?: string
  merge_reason?: string
  close_reason?: string
}

interface Diff {
  file_path: string
  diff_text: string
}

interface Comment {
  github_id: number
  body: string
  author: string
  created_at: string
  updated_at: string
}

interface DetailData {
  id: number
  title: string
  body: string
  state: string
  author: string
  created_at: string
  updated_at: string
  url: string
  summary?: Summary
  diffs?: Diff[]
  comments?: Comment[]
}

export default function Detail({ type }: DetailProps) {
  const { id } = useParams<{ id: string }>()
  const [data, setData] = useState<DetailData | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'summary' | 'diffs' | 'comments' | 'analysis'>('summary')

  useEffect(() => {
    loadData()
  }, [id])

  const loadData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const result = type === 'pr' ? await getPR(parseInt(id)) : await getIssue(parseInt(id))
      setData(result as DetailData)
    } catch (error) {
      console.error('データの取得に失敗しました', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="container mx-auto px-4 py-8">読み込み中...</div>
  }

  if (!data) {
    return <div className="container mx-auto px-4 py-8">データが見つかりません</div>
  }

  const hasSummary = data.summary && Object.keys(data.summary).length > 0
  const hasDiffs = data.diffs && data.diffs.length > 0
  const hasComments = data.comments && data.comments.length > 0

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <Link to="/" className="text-blue-600 hover:underline mb-4 inline-block">
          ← ダッシュボードに戻る
        </Link>

        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex items-center gap-2 mb-4">
            <StatusBadge state={data.state} />
            <h1 className="text-2xl font-bold text-gray-900">{data.title}</h1>
          </div>

          <div className="prose max-w-none mb-6">
            <p className="text-gray-700 whitespace-pre-wrap">{data.body}</p>
          </div>

          <div className="border-t border-gray-200 pt-4 mt-6">
            <div className="text-sm text-gray-600">
              <p>作成者: {data.author}</p>
              <p>作成日時: {new Date(data.created_at).toLocaleString('ja-JP')}</p>
              <p>更新日時: {new Date(data.updated_at).toLocaleString('ja-JP')}</p>
              {data.url && (
                <a href={data.url} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                  GitHubで開く
                </a>
              )}
            </div>
          </div>
        </div>

        {/* タブナビゲーション */}
        <div className="bg-white rounded-lg shadow mb-6">
          <div className="border-b border-gray-200">
            <nav className="flex -mb-px">
              {hasSummary && (
                <button
                  onClick={() => setActiveTab('summary')}
                  className={`px-6 py-3 text-sm font-medium ${
                    activeTab === 'summary'
                      ? 'border-b-2 border-blue-500 text-blue-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  要約
                </button>
              )}
              {type === 'pr' && hasDiffs && (
                <button
                  onClick={() => setActiveTab('diffs')}
                  className={`px-6 py-3 text-sm font-medium ${
                    activeTab === 'diffs'
                      ? 'border-b-2 border-blue-500 text-blue-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  差分
                </button>
              )}
              {hasComments && (
                <button
                  onClick={() => setActiveTab('comments')}
                  className={`px-6 py-3 text-sm font-medium ${
                    activeTab === 'comments'
                      ? 'border-b-2 border-blue-500 text-blue-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  コメント ({data.comments?.length || 0})
                </button>
              )}
              {type === 'pr' && (data.state === 'merged' || data.state === 'closed') && (
                <button
                  onClick={() => setActiveTab('analysis')}
                  className={`px-6 py-3 text-sm font-medium ${
                    activeTab === 'analysis'
                      ? 'border-b-2 border-blue-500 text-blue-600'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  分析結果
                </button>
              )}
            </nav>
          </div>

          <div className="p-6">
            {/* 要約タブ */}
            {activeTab === 'summary' && hasSummary && (
              <div className="space-y-6">
                {data.summary?.description_summary && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">説明の要約</h3>
                    <div className="bg-gray-50 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.description_summary}</p>
                    </div>
                  </div>
                )}

                {type === 'pr' && data.summary?.diff_summary && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">差分の要約</h3>
                    <div className="bg-gray-50 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.diff_summary}</p>
                    </div>
                  </div>
                )}

                {type === 'pr' && data.summary?.diff_explanation && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">差分の解説</h3>
                    <div className="bg-gray-50 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.diff_explanation}</p>
                    </div>
                  </div>
                )}

                {data.summary?.comments_summary && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">コメントの要約</h3>
                    <div className="bg-gray-50 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.comments_summary}</p>
                    </div>
                  </div>
                )}

                {data.summary?.discussion_summary && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">議論の要約</h3>
                    <div className="bg-gray-50 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.discussion_summary}</p>
                    </div>
                  </div>
                )}

                {!hasSummary && (
                  <div className="text-center py-8 text-gray-500">
                    要約データがまだ生成されていません。しばらくお待ちください。
                  </div>
                )}
              </div>
            )}

            {/* 差分タブ */}
            {activeTab === 'diffs' && type === 'pr' && (
              <div className="space-y-4">
                {hasDiffs ? (
                  data.diffs?.map((diff, index) => (
                    <div key={index} className="border border-gray-200 rounded-lg overflow-hidden">
                      <div className="bg-gray-100 px-4 py-2 border-b border-gray-200">
                        <h4 className="text-sm font-semibold text-gray-900">{diff.file_path}</h4>
                      </div>
                      <div className="p-4">
                        <pre className="text-sm overflow-x-auto">
                          <code className="whitespace-pre-wrap">{diff.diff_text}</code>
                        </pre>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    差分データがありません。
                  </div>
                )}
              </div>
            )}

            {/* コメントタブ */}
            {activeTab === 'comments' && (
              <div className="space-y-4">
                {hasComments ? (
                  data.comments?.map((comment) => (
                    <div key={comment.github_id} className="border border-gray-200 rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-semibold text-gray-900">{comment.author}</span>
                        <span className="text-sm text-gray-500">
                          {new Date(comment.created_at).toLocaleString('ja-JP')}
                        </span>
                      </div>
                      <div className="prose max-w-none">
                        <p className="text-gray-700 whitespace-pre-wrap">{comment.body}</p>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    コメントがありません。
                  </div>
                )}
              </div>
            )}

            {/* 分析結果タブ */}
            {activeTab === 'analysis' && type === 'pr' && (
              <div className="space-y-6">
                {data.state === 'merged' && data.summary?.merge_reason && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">Merge理由の分析</h3>
                    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.merge_reason}</p>
                    </div>
                  </div>
                )}

                {data.state === 'closed' && data.summary?.close_reason && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">Close理由の分析</h3>
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                      <p className="text-gray-700 whitespace-pre-wrap">{data.summary.close_reason}</p>
                    </div>
                  </div>
                )}

                {!data.summary?.merge_reason && !data.summary?.close_reason && (
                  <div className="text-center py-8 text-gray-500">
                    分析結果がまだ生成されていません。
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
