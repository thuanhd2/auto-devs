import { useMemo } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'

interface PlanPreviewProps {
  content: string
  className?: string
  showActions?: boolean
  printFriendly?: boolean
}

// Basic markdown parser for common elements
const parseMarkdown = (content: string): string => {
  if (!content) return ''

  let html = content
    // Headers
    .replace(
      /^### (.*$)/gm,
      '<h4 class="text-lg font-semibold mt-6 mb-3 text-gray-900">$1</h3>'
    )
    .replace(
      /^## (.*$)/gm,
      '<h3 class="text-xl font-semibold mt-8 mb-4 text-gray-900 border-b border-gray-200 pb-2">$1</h2>'
    )
    .replace(
      /^# (.*$)/gm,
      '<h2 class="text-2xl font-bold mt-8 mb-6 text-gray-900">$1</h1>'
    )

    // Bold and italic
    .replace(
      /\*\*\*(.*?)\*\*\*/g,
      '<strong><em class="font-bold italic">$1</em></strong>'
    )
    .replace(/\*\*(.*?)\*\*/g, '<strong class="font-semibold">$1</strong>')
    .replace(/\*(.*?)\*/g, '<em class="italic">$1</em>')

    // Code blocks
    .replace(
      /```(\w+)?\n([\s\S]*?)\n```/g,
      '<pre class="bg-gray-100 border rounded-lg p-4 text-sm font-mono overflow-x-auto my-4"><code class="language-$1">$2</code></pre>'
    )

    // Inline code
    .replace(
      /`([^`]+)`/g,
      '<code class="bg-gray-100 px-1.5 py-0.5 rounded text-sm font-mono">$1</code>'
    )

    // Lists
    .replace(/^\* (.*$)/gm, '<li class="ml-4">$1</li>')
    .replace(/^- (.*$)/gm, '<li class="ml-4">$1</li>')
    .replace(/^\d+\. (.*$)/gm, '<li class="ml-4">$1</li>')

    // Blockquotes
    .replace(
      /^> (.*$)/gm,
      '<blockquote class="border-l-4 border-blue-500 pl-4 my-4 text-gray-700 italic">$1</blockquote>'
    )

    // Links
    .replace(
      /\[([^\]]+)\]\(([^)]+)\)/g,
      '<a href="$2" class="text-blue-600 hover:text-blue-800 underline" target="_blank" rel="noopener noreferrer">$1</a>'
    )

    // Line breaks
    .replace(/\n\n/g, '</p><p class="mb-4">')
    .replace(/\n/g, '<br>')

  // Wrap in paragraph tags and handle lists
  html = '<p class="mb-4">' + html + '</p>'

  // Fix list formatting
  html = html.replace(
    /<\/p><p class="mb-4">(<li class="ml-4">.*?)<\/p>/g,
    '<ul class="list-disc list-inside space-y-1 my-4">$1</ul>'
  )
  html = html.replace(
    /(<li class="ml-4">.*?)<br>(<li class="ml-4">)/g,
    '$1</li>$2'
  )
  html = html.replace(/(<li class="ml-4">.*?)(<\/ul>)/g, '$1</li>$2')

  return html
}

export function PlanPreview({
  content,
  className = '',
  showActions = true,
  printFriendly = false,
}: PlanPreviewProps) {
  const parsedContent = useMemo(() => parseMarkdown(content), [content])

  const handleCopyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(content)
      toast.success('Plan copied to clipboard!')
    } catch (error) {
      toast.error('Failed to copy to clipboard')
    }
  }

  const handlePrint = () => {
    const printWindow = window.open('', '_blank')
    if (!printWindow) return

    const printContent = `
      <!DOCTYPE html>
      <html>
        <head>
          <title>Implementation Plan</title>
          <style>
            body { 
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              line-height: 1.6;
              max-width: 800px;
              margin: 0 auto;
              padding: 20px;
              color: #333;
            }
            h1, h2, h3 { color: #1a1a1a; margin-top: 1.5em; }
            pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
            code { background: #f0f0f0; padding: 2px 4px; border-radius: 3px; }
            blockquote { border-left: 4px solid #007acc; padding-left: 15px; margin: 15px 0; }
            ul, ol { padding-left: 20px; }
            a { color: #007acc; }
            @media print {
              body { print-color-adjust: exact; }
            }
          </style>
        </head>
        <body>
          <div>
            ${parsedContent}
          </div>
        </body>
      </html>
    `

    printWindow.document.open()
    printWindow.document.write(printContent)
    printWindow.document.close()
    printWindow.print()
  }

  const handleOpenInNewTab = () => {
    const newWindow = window.open('', '_blank')
    if (!newWindow) return

    const fullContent = `
      <!DOCTYPE html>
      <html>
        <head>
          <title>Implementation Plan</title>
          <meta name="viewport" content="width=device-width, initial-scale=1">
          <style>
            body { 
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              line-height: 1.6;
              max-width: 900px;
              margin: 0 auto;
              padding: 20px;
              color: #333;
              background: #fff;
            }
            h1 { font-size: 2em; font-weight: bold; margin: 2em 0 1.5em 0; color: #1a1a1a; }
            h2 { font-size: 1.5em; font-weight: 600; margin: 2em 0 1em 0; color: #1a1a1a; border-bottom: 2px solid #e5e5e5; padding-bottom: 0.5em; }
            h3 { font-size: 1.25em; font-weight: 600; margin: 1.5em 0 0.75em 0; color: #1a1a1a; }
            p { margin-bottom: 1em; }
            pre { 
              background: #f8f9fa; 
              padding: 15px; 
              border-radius: 8px; 
              overflow-x: auto; 
              border: 1px solid #e9ecef;
              margin: 1.5em 0;
            }
            code { 
              background: #f1f3f4; 
              padding: 2px 6px; 
              border-radius: 4px; 
              font-size: 0.9em;
              font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', monospace;
            }
            blockquote { 
              border-left: 4px solid #007acc; 
              padding-left: 15px; 
              margin: 1.5em 0; 
              color: #666;
              font-style: italic;
            }
            ul, ol { padding-left: 20px; margin: 1em 0; }
            li { margin: 0.5em 0; }
            a { color: #007acc; text-decoration: underline; }
            a:hover { color: #005a9e; }
          </style>
        </head>
        <body>
          <div>
            ${parsedContent}
          </div>
        </body>
      </html>
    `

    newWindow.document.open()
    newWindow.document.write(fullContent)
    newWindow.document.close()
  }

  if (!content.trim()) {
    return (
      <Card className={`p-8 text-center ${className}`}>
        <div className='mb-4 text-gray-400'>
          <svg
            className='mx-auto h-12 w-12'
            fill='none'
            viewBox='0 0 24 24'
            stroke='currentColor'
          >
            <path
              strokeLinecap='round'
              strokeLinejoin='round'
              strokeWidth={2}
              d='M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
            />
          </svg>
        </div>
        <h3 className='mb-2 text-lg font-medium text-gray-900'>
          No content to preview
        </h3>
        <p className='text-sm text-gray-500'>
          Start writing your plan to see a preview here.
        </p>
      </Card>
    )
  }

  return (
    <div className={`${className}`}>
      {/* Actions Bar */}
      {showActions && !printFriendly && (
        <div className='mb-4 flex items-center justify-between rounded-lg bg-gray-50 p-3'>
          <div className='flex items-center gap-2'>
            <Button
              variant='ghost'
              size='sm'
              onClick={handleOpenInNewTab}
              className='h-8'
            >
              <ExternalLink className='mr-1 h-4 w-4' />
              Full Screen
            </Button>
            {/* <Button
              variant='ghost'
              size='sm'
              onClick={handlePrint}
              className='h-8'
            >
              <Printer className='mr-1 h-4 w-4' />
              Print
            </Button> */}
          </div>
        </div>
      )}

      {/* Content */}
      <div
        className={`prose prose-sm max-w-none ${printFriendly ? 'print:prose-print' : ''}`}
      >
        <div
          className='leading-relaxed'
          dangerouslySetInnerHTML={{ __html: parsedContent }}
        />
      </div>
    </div>
  )
}
