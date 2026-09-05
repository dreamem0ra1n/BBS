export const CONTENT_MAX_LENGTH = 5000

export function contentLength(value) {
  return Array.from(value || '').length
}

export function limitContent(value, maxLength = CONTENT_MAX_LENGTH) {
  return Array.from(value || '')
    .slice(0, maxLength)
    .join('')
}
