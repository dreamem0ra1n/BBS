function encode(value: unknown): string {
  return encodeURIComponent(value === undefined || value === null ? '' : String(value))
}

export function encodeForm(data: Record<string, unknown> = {}): string {
  return Object.keys(data)
    .filter((key) => data[key] !== undefined && data[key] !== null)
    .map((key) => `${encode(key)}=${encode(data[key])}`)
    .join('&')
}
