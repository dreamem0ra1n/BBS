export default function hideFile(string) {
  console.log(string.replace(fileExp, ''))
  return string.replace(fileExp, '')
}
export function generateFileString(FileIds) {
  return (
    '[File]' +
    FileIds.reduce(
      (string, id) => {
        return string + ' ' + id.toString()
      },
      ['']
    ) +
    '[/File]'
  )
}
export function getFileIds(content) {
  let FileString = content.match(fileExp)
  if (!FileString) return []
  FileString = FileString[0]
  FileString = FileString.replace('[File]', '').replace('[/File]', '')
  const FileIds = FileString.split(' ')
  return FileIds
}
export const fileExp = /\[File\].+\[\/File\]/
