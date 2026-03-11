import { FormField } from '../ui/FormField'

/**
 * HouseholdSetupCard handles first household creation.
 */
export function HouseholdSetupCard({
  newHouseholdName,
  onHouseholdNameChange,
  newSettlementMode,
  onSettlementModeChange,
  newCurrency,
  onCurrencyChange,
  onSubmit,
  isSubmitting,
  isLoading,
}) {
  return (
    <section className="card" aria-label="Create first household">
      <h2 className="sectionTitle">Create your first household</h2>
      <form className="formStack" onSubmit={onSubmit}>
        <FormField label="Name" htmlFor="newHouseholdName">
          <input
            id="newHouseholdName"
            className="input"
            placeholder="e.g. Casa Familia"
            value={newHouseholdName}
            onChange={(e) => onHouseholdNameChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          />
        </FormField>
        <FormField label="Settlement mode" htmlFor="newSettlementMode">
          <select
            id="newSettlementMode"
            className="input"
            value={newSettlementMode}
            onChange={(e) => onSettlementModeChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          >
            <option value="equal">Equal split</option>
            <option value="proportional">Proportional to salary</option>
          </select>
        </FormField>
        <FormField label="Currency" htmlFor="newCurrency">
          <input
            id="newCurrency"
            className="input"
            placeholder="MXN"
            value={newCurrency}
            onChange={(e) => onCurrencyChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          />
        </FormField>
        <button type="submit" className="btn btnPrimary" disabled={isSubmitting || isLoading || !newHouseholdName.trim()}>
          {isSubmitting ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Creating...</> : 'Create household'}
        </button>
      </form>
    </section>
  )
}
