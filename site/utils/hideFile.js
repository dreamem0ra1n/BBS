export default function hideFile(string) {
  console.log(string.replace(fileExp, ''))
  return string.replace(fileExp, '')
}
export const fileExp = /\[File\].+\[\/File\]/
