import { FormField } from '../ui/FormField'

/**
 * MemberSetupCard handles first member creation.
 */
export function MemberSetupCard({
  newMemberName,
  onMemberNameChange,
  newMemberEmail,
  onMemberEmailChange,
  newMemberSalary,
  onMemberSalaryChange,
  onSubmit,
  isSubmitting,
  isLoading,
}) {
  return (
    <section className="card" aria-label="Create first member">
      <h2 className="sectionTitle">Create your first member</h2>
      <form className="formStack" onSubmit={onSubmit}>
        <FormField label="Name" htmlFor="newMemberName">
          <input
            id="newMemberName"
            className="input"
            placeholder="e.g. Alex"
            value={newMemberName}
            onChange={(e) => onMemberNameChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          />
        </FormField>
        <FormField label="Email" htmlFor="newMemberEmail">
          <input
            id="newMemberEmail"
            className="input"
            type="email"
            placeholder="alex@example.com"
            value={newMemberEmail}
            onChange={(e) => onMemberEmailChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          />
        </FormField>
        <FormField label="Monthly salary (cents)" htmlFor="newMemberSalary">
          <input
            id="newMemberSalary"
            className="input"
            type="number"
            min="0"
            placeholder="0"
            value={newMemberSalary}
            onChange={(e) => onMemberSalaryChange(e.target.value)}
            disabled={isSubmitting || isLoading}
          />
        </FormField>
        <button
          type="submit"
          className="btn btnPrimary"
          disabled={isSubmitting || isLoading || !newMemberName.trim() || !newMemberEmail.trim()}
        >
          {isSubmitting ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Creating...</> : 'Create member'}
        </button>
      </form>
    </section>
  )
}
