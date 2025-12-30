import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getPRs, getIssues, sync } from '../services/api'
import StatusBadge from '../components/StatusBadge'
import ManualSyncButton from '../components/ManualSyncButton'

export default function Dashboard() {
  const [prs, setPRs] = useState<any[]>([])
  const [issues, setIssues] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState<'prs' | 'issues'>('prs')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [prsData, issuesData] = await Promise.all([
        getPRs({ limit: 50 }),
        getIssues({ limit: 50 }),
      ])
      setPRs(prsData.data || [])
      setIssues(issuesData.data || [])
    } catch (error) {
      console.error('データの取得に失敗しました', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSync = async () => {
    try {
      await sync()
      setTimeout(loadData, 2000) // 2秒後に再読み込み
    } catch (error) {
      console.error('同期に失敗しました', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-gray-900">UpGo - Goリポジトリ監視システム</h1>
          <ManualSyncButton onSync={handleSync} />
        </div>

        <div className="bg-white rounded-lg shadow">
          <div className="border-b border-gray-200">
            <nav className="flex -mb-px">
              <button
                onClick={() => setActiveTab('prs')}
                className={`px-6 py-3 text-sm font-medium ${
                  activeTab === 'prs'
                    ? 'border-b-2 border-blue-500 text-blue-600'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                Pull Requests
              </button>
              <button
                onClick={() => setActiveTab('issues')}
                className={`px-6 py-3 text-sm font-medium ${
                  activeTab === 'issues'
                    ? 'border-b-2 border-blue-500 text-blue-600'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                Issues
              </button>
            </nav>
          </div>

          <div className="p-6">
            {loading ? (
              <div className="text-center py-8">読み込み中...</div>
            ) : activeTab === 'prs' ? (
              <div className="space-y-4">
                {prs.map((pr) => (
                  <Link
                    key={pr.id}
                    to={`/pr/${pr.id}`}
                    className="block p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <StatusBadge state={pr.state} />
                          <h3 className="text-lg font-semibold text-gray-900">{pr.title}</h3>
                        </div>
                        <p className="text-sm text-gray-600 mb-2">{pr.body?.substring(0, 200)}...</p>
                        <div className="text-xs text-gray-500">
                          by {pr.author} • {new Date(pr.created_at).toLocaleDateString('ja-JP')}
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="space-y-4">
                {issues.map((issue) => (
                  <Link
                    key={issue.id}
                    to={`/issue/${issue.id}`}
                    className="block p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <StatusBadge state={issue.state} />
                          <h3 className="text-lg font-semibold text-gray-900">{issue.title}</h3>
                        </div>
                        <p className="text-sm text-gray-600 mb-2">{issue.body?.substring(0, 200)}...</p>
                        <div className="text-xs text-gray-500">
                          by {issue.author} • {new Date(issue.created_at).toLocaleDateString('ja-JP')}
                        </div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
