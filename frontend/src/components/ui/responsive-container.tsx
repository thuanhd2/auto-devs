import * as React from 'react'
import { cn } from '@/lib/utils'

interface ResponsiveContainerProps {
  children: React.ReactNode
  className?: string
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full'
  padding?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
  center?: boolean
}

const maxWidthClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md', 
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
  full: 'max-w-full',
}

const paddingClasses = {
  none: '',
  sm: 'px-2 sm:px-4',
  md: 'px-4 sm:px-6',
  lg: 'px-6 sm:px-8',
  xl: 'px-8 sm:px-12',
}

export function ResponsiveContainer({
  children,
  className,
  maxWidth = 'full',
  padding = 'md',
  center = true,
}: ResponsiveContainerProps) {
  return (
    <div
      className={cn(
        'w-full',
        maxWidthClasses[maxWidth],
        paddingClasses[padding],
        center && 'mx-auto',
        className
      )}
    >
      {children}
    </div>
  )
}

// Responsive Grid Component
interface ResponsiveGridProps {
  children: React.ReactNode
  className?: string
  cols?: {
    default?: 1 | 2 | 3 | 4 | 5 | 6
    sm?: 1 | 2 | 3 | 4 | 5 | 6
    md?: 1 | 2 | 3 | 4 | 5 | 6
    lg?: 1 | 2 | 3 | 4 | 5 | 6
    xl?: 1 | 2 | 3 | 4 | 5 | 6
  }
  gap?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
}

const colClasses = {
  1: 'grid-cols-1',
  2: 'grid-cols-2',
  3: 'grid-cols-3',
  4: 'grid-cols-4',
  5: 'grid-cols-5',
  6: 'grid-cols-6',
}

const gapClasses = {
  none: 'gap-0',
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
  xl: 'gap-8',
}

export function ResponsiveGrid({
  children,
  className,
  cols = { default: 1, sm: 2, md: 3, lg: 4 },
  gap = 'md',
}: ResponsiveGridProps) {
  const gridClasses = cn(
    'grid',
    cols.default && colClasses[cols.default],
    cols.sm && `sm:${colClasses[cols.sm]}`,
    cols.md && `md:${colClasses[cols.md]}`,
    cols.lg && `lg:${colClasses[cols.lg]}`,
    cols.xl && `xl:${colClasses[cols.xl]}`,
    gapClasses[gap],
    className
  )

  return <div className={gridClasses}>{children}</div>
}

// Responsive Stack Component
interface ResponsiveStackProps {
  children: React.ReactNode
  className?: string
  direction?: {
    default?: 'row' | 'column'
    sm?: 'row' | 'column'
    md?: 'row' | 'column'
    lg?: 'row' | 'column'
  }
  align?: 'start' | 'center' | 'end' | 'stretch' | 'baseline'
  justify?: 'start' | 'center' | 'end' | 'between' | 'around' | 'evenly'
  gap?: 'none' | 'sm' | 'md' | 'lg' | 'xl'
  wrap?: boolean
}

const directionClasses = {
  row: 'flex-row',
  column: 'flex-col',
}

const alignClasses = {
  start: 'items-start',
  center: 'items-center',
  end: 'items-end',
  stretch: 'items-stretch',
  baseline: 'items-baseline',
}

const justifyClasses = {
  start: 'justify-start',
  center: 'justify-center',
  end: 'justify-end',
  between: 'justify-between',
  around: 'justify-around',
  evenly: 'justify-evenly',
}

export function ResponsiveStack({
  children,
  className,
  direction = { default: 'column', md: 'row' },
  align = 'start',
  justify = 'start',
  gap = 'md',
  wrap = false,
}: ResponsiveStackProps) {
  const stackClasses = cn(
    'flex',
    direction.default && directionClasses[direction.default],
    direction.sm && `sm:${directionClasses[direction.sm]}`,
    direction.md && `md:${directionClasses[direction.md]}`,
    direction.lg && `lg:${directionClasses[direction.lg]}`,
    alignClasses[align],
    justifyClasses[justify],
    gapClasses[gap],
    wrap && 'flex-wrap',
    className
  )

  return <div className={stackClasses}>{children}</div>
}

// Responsive Show/Hide Component
interface ResponsiveVisibilityProps {
  children: React.ReactNode
  show?: {
    default?: boolean
    sm?: boolean
    md?: boolean
    lg?: boolean
    xl?: boolean
  }
  hide?: {
    default?: boolean
    sm?: boolean
    md?: boolean
    lg?: boolean
    xl?: boolean
  }
}

export function ResponsiveVisibility({
  children,
  show,
  hide,
}: ResponsiveVisibilityProps) {
  const visibilityClasses = cn(
    // Show classes
    show?.default === false && 'hidden',
    show?.default === true && 'block',
    show?.sm === false && 'sm:hidden',
    show?.sm === true && 'sm:block',
    show?.md === false && 'md:hidden',
    show?.md === true && 'md:block',
    show?.lg === false && 'lg:hidden',
    show?.lg === true && 'lg:block',
    show?.xl === false && 'xl:hidden',
    show?.xl === true && 'xl:block',
    
    // Hide classes
    hide?.default === true && 'hidden',
    hide?.sm === true && 'sm:hidden',
    hide?.md === true && 'md:hidden',
    hide?.lg === true && 'lg:hidden',
    hide?.xl === true && 'xl:hidden',
  )

  return <div className={visibilityClasses}>{children}</div>
}

// Responsive Text Component
interface ResponsiveTextProps {
  children: React.ReactNode
  className?: string
  size?: {
    default?: 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl'
    sm?: 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl'
    md?: 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl'
    lg?: 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl'
  }
  weight?: 'light' | 'normal' | 'medium' | 'semibold' | 'bold'
  align?, center?: 'left' | 'center' | 'right'
  as?: 'p' | 'span' | 'div' | 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
}

const sizeClasses = {
  xs: 'text-xs',
  sm: 'text-sm',
  base: 'text-base',
  lg: 'text-lg',
  xl: 'text-xl',
  '2xl': 'text-2xl',
  '3xl': 'text-3xl',
  '4xl': 'text-4xl',
}

const weightClasses = {
  light: 'font-light',
  normal: 'font-normal',
  medium: 'font-medium',
  semibold: 'font-semibold',
  bold: 'font-bold',
}

const alignClasses = {
  left: 'text-left',
  center: 'text-center',
  right: 'text-right',
}

export function ResponsiveText({
  children,
  className,
  size = { default: 'base' },
  weight = 'normal',
  align = 'left',
  as: Component = 'p',
}: ResponsiveTextProps) {
  const textClasses = cn(
    size.default && sizeClasses[size.default],
    size.sm && `sm:${sizeClasses[size.sm]}`,
    size.md && `md:${sizeClasses[size.md]}`,
    size.lg && `lg:${sizeClasses[size.lg]}`,
    weightClasses[weight],
    alignClasses[align],
    className
  )

  return <Component className={textClasses}>{children}</Component>
}