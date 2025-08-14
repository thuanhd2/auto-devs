import { IconBrandGithub } from '@tabler/icons-react'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as SearchComponent } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

export default function AboutUs() {
  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <div className='ml-auto flex items-center space-x-4'>
          <SearchComponent />
          <ThemeSwitch />
        </div>
      </Header>
      <Main>
        <div className='space-y-0.5'>
          <h1 className='text-2xl font-bold tracking-tight'>
            About This project
          </h1>
          <p className='text-muted-foreground'>
            Nothing rather than a project to learn and practice.
          </p>
        </div>
        <Separator className='my-4 lg:my-6' />
        <a
          href='https://github.com/thuanhd2/auto-devs'
          target='_blank'
          rel='noopener noreferrer'
        >
          <Button>
            <IconBrandGithub />
            Github Repository
          </Button>
        </a>
      </Main>
    </>
  )
}
