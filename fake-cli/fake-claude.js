import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const filePath = 'fake-output.log'
const absolutePath = path.join(__dirname, filePath)

const dummyOutput = fs.readFileSync(absolutePath, 'utf8')

async function main() {
    const lines = dummyOutput.split('\n')
    for (const line of lines) {
        console.log(line)
        await new Promise(resolve => setTimeout(resolve, 300))
    }
    // dummy some files in the current directory
    const files = ['file1.txt', 'file2.txt', 'file3.txt']
    for (const file of files) {
        fs.writeFileSync(file, 'dummy content - ' + new Date().toISOString())
    }
}

main()
