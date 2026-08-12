export default function HTMLDecode(text) {
  try {
    let temp = document.createElement('div')
    temp.innerHTML = text
    const output = temp.innerText || temp.textContent
    temp = null
    return output
  } catch (e) {
    console.log(e)
    return text
  }
}
