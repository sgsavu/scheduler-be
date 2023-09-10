export const downloadBlob = (blob: Blob, filename: string) => {
     const a = document.createElement("a")
     const url = URL.createObjectURL(blob)

     const ext = blob.type.split('/')[1]

     a.href = url
     a.download = filename + "." + ext
     document.body.appendChild(a)
     a.click()

     setTimeout(() => {
       document.body.removeChild(a);
       window.URL.revokeObjectURL(url);
     }, 0)
}