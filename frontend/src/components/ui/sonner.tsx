import { Toaster as Sonner, ToasterProps } from 'sonner'
import { useTheme } from '@/context/theme-context'

const const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = 'system' } = useTheme()

  return (
    <Sonner
      theme={theme as ToasterProps['theme']}
      className='toaster group [&_div[data-content]]:w-full'
      style={
        {
          '--normal-bg': 'var(--popover)',
          '--normal-text': 'var(--popover-foreground)',
          '--normal-border': 'var(--border)',
        } as React.CSSProperties
      }
      position="top-right"
      expand={true}
      richColors={true}
      closeButton={true}
      toastOptions={{
        style: {
          background: 'var(--background)',
          border: '1px solid var(--border)',
          color: 'var(--foreground)',
        },
        className: 'group toast group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg',
        descriptionClassName: 'group-[.toast]:text-muted-foreground',
        actionButtonClassName: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
        cancelButtonClassName: 'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground',
      }}
      offset={16}
      gap={8}
      visibleToasts={5}
      duration={4000}
      {...props}
    />
  )
}
      {...props}
    />
  )
}

export { Toaster }
