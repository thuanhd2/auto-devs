import { config } from '../config.js';

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogContext {
  component?: string;
  requestId?: string;
  [key: string]: unknown;
}

class Logger {
  private context: LogContext = {};

  setContext(ctx: LogContext) {
    this.context = { ...this.context, ...ctx };
  }

  private formatMessage(level: LogLevel, message: string, data?: unknown): string {
    const timestamp = new Date().toISOString();
    const component = this.context.component ? `[${this.context.component}]` : '';
    const contextStr = this.context.requestId ? ` (${this.context.requestId})` : '';

    let output = `${timestamp} ${level.toUpperCase()} ${component}${contextStr} ${message}`;

    if (data && config.debug) {
      output += `\n${JSON.stringify(data, null, 2)}`;
    }

    return output;
  }

  debug(message: string, data?: unknown) {
    if (config.debug) {
      console.error(this.formatMessage('debug', message, data));
    }
  }

  info(message: string, data?: unknown) {
    console.error(this.formatMessage('info', message, data));
  }

  warn(message: string, data?: unknown) {
    console.error(this.formatMessage('warn', message, data));
  }

  error(message: string, data?: unknown) {
    console.error(this.formatMessage('error', message, data));
  }
}

export const logger = new Logger();
