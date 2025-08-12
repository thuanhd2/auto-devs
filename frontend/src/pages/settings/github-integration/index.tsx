import ContentSection from '../components/content-section'
import GithubIntegrationForm from './github-integration-form'

export default function SettingsProfile() {
  return (
    <ContentSection
      title='Github Integration'
      desc='The github Personal Access Token is now in environment variables, it should will be configurable in here soon...'
    >
      <GithubIntegrationForm />
    </ContentSection>
  )
}
