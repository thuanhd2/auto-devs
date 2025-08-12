import ContentSection from '../components/content-section'
import { AiExecutorForm } from './ai-executor'

export default function AiExecutor() {
  const description =
    `You can choose your default AI executor, it should be one of Claude Code, ` +
    `Codex, Google Gemini, or Qween Code... and/or its MCP servers configurations.`
  return (
    <ContentSection title='AI Executor' desc={description}>
      <AiExecutorForm />
    </ContentSection>
  )
}
