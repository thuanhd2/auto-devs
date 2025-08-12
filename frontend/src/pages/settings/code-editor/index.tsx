import ContentSection from '../components/content-section'
import { CodeEditorForm } from './code-editor-form'

export default function CodeEditor() {
  return (
    <ContentSection
      title='Code Editor'
      desc='Maybe Vscode, Cursor or Windows Terminal?'
    >
      <CodeEditorForm />
    </ContentSection>
  )
}
