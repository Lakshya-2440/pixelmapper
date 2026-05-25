import { Copy } from 'lucide-react'

export default function EventsTable({ events, loading, onCopy, copied }) {
  if (loading) {
    return <div className="table-state">Loading events...</div>
  }

  if (events.length === 0) {
    return <div className="table-state">No events match current filters.</div>
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Token</th>
            <th>Pixel ID</th>
            <th>User</th>
            <th>IP</th>
            <th>User Agent</th>
            <th>Timestamp</th>
            <th aria-label="Actions" />
          </tr>
        </thead>
        <tbody>
          {events.map((event) => (
            <tr key={event.id}>
              <td>
                <div className="stacked">
                  <strong>{event.token}</strong>
                  <span>{event.label || 'Untitled'}</span>
                </div>
              </td>
              <td className="mono">{event.pixel_id}</td>
              <td>{event.uid || event.email || '-'}</td>
              <td>{event.ip || '-'}</td>
              <td className="ua">{event.user_agent || '-'}</td>
              <td>{formatDate(event.created_at)}</td>
              <td>
                <button className="icon-button" type="button" onClick={() => onCopy(`/t/${event.token}`)} title="Copy path" aria-label="Copy path">
                  <Copy size={16} />
                </button>
                <span className="visually-hidden">{copied === `/t/${event.token}` ? 'Copied' : ''}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function formatDate(value) {
  if (!value) return '-'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
