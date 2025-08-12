import ContentSection from '../components/content-section'
import { AiExecutorForm } from './ai-executor'

export default function AiExecutor() {
  return (
    <ContentSection
      title='AI Executor'
      desc='You can choose your default AI executor, it should be one of Claude Code, Codex, Google Gemini, or Qween Code....'
    >
      <AiExecutorForm />
    </ContentSection>
  )
}
