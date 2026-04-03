import { Link } from 'react-router-dom'

/**
 * AppHeader — top bar with brand identity, household selector, reload
 * action, and backend health indicator.
 */
export function AppHeader({ health, householdId, onHouseholdChange, onReload, onLogout, isLoading, households = [] }) {
  const isLive = health === 'ok'

  return (
    <header className="appHeader">
      {/* Brand */}
      <div className="brand">
        <div className="brandMark" aria-hidden />
        <div className="brandMeta">
          <div className="brandEyebrow">Personal Ledger</div>
          <div className="brandName">micha</div>
          <div className="brandTagline">Household expense tracker</div>
        </div>
      </div>

      {/* Controls */}
      <div className="headerControls">
        {/* Household selector */}
        <div className="householdRow">
          <label htmlFor="householdInput" className="householdLabel">
            Household
          </label>
          <select
            id="householdInput"
            className="householdInput"
            value={householdId}
            onChange={(e) => onHouseholdChange(e.target.value)}
            aria-label="Household ID"
          >
            <option value="">Select household</option>
            {households.map((household) => (
              <option key={household.id} value={household.id}>
                {household.name}
              </option>
            ))}
          </select>
        </div>

        {/* Invite member */}
        {householdId && (
          <Link
            to="/members/new"
            className="btn btnGhost btnSm"
            aria-label="Invite a new member"
          >
            + Member
          </Link>
        )}

        {householdId && (
          <Link
            to="/household/settings"
            className="btn btnGhost btnSm"
            aria-label="Open household settings"
          >
            Settings
          </Link>
        )}

        {/* Reload */}
        <button
          type="button"
          className="btn btnGhost btnSm"
          onClick={onReload}
          disabled={isLoading}
          aria-label="Reload expenses"
        >
          {isLoading && <span className="spinIcon" aria-hidden />}
          {isLoading ? 'Loading…' : 'Reload'}
        </button>

        <button
          type="button"
          className="btn btnGhostDanger btnSm"
          onClick={() => onLogout()}
          disabled={isLoading}
          aria-label="Sign out"
        >
          Sign out
        </button>

        {/* Health */}
        <span className={isLive ? 'pill pillOk' : 'pill pillOff'} aria-label={`Backend status: ${health}`}>
          {isLive ? 'live' : health}
        </span>
      </div>
    </header>
  )
}
