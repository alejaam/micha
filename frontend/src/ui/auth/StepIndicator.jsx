/**
 * StepIndicator — horizontal dots/labels for onboarding wizard.
 *
 * @param {{ path: string, label: string }[]} steps
 * @param {string} currentPath
 */
export function StepIndicator({ steps, currentPath }) {
    const currentIndex = steps.findIndex((s) => s.path === currentPath)
    const displayIndex = currentIndex >= 0 ? currentIndex + 1 : 1

    return (
        <nav className="pd-steps" aria-label={`Paso ${displayIndex} de ${steps.length}`}>
            {steps.map((step, i) => {
                const isActive = step.path === currentPath
                const isCompleted = i < currentIndex
                const dotClass = [
                    'pd-stepDot',
                    isActive ? 'pd-stepDotActive' : '',
                    isCompleted ? 'pd-stepDotCompleted' : '',
                ].filter(Boolean).join(' ')

                return (
                    <div
                        key={step.path}
                        className={`pd-step ${isActive ? 'pd-stepActive' : ''}`}
                    >
                        <span className={dotClass} />
                        <span className="pd-stepLabel">{step.label}</span>
                    </div>
                )
            })}
        </nav>
    )
}
