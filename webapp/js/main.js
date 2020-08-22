import 'bulma/css/bulma.css'
import * as monaco from 'monaco-editor'
import 'whatwg-fetch'

self.MonacoEnvironment = {
  getWorkerUrl: function () {
    return './editor.worker.js'
  }
}

// Register a new language
monaco.languages.register({ id: 'goink' })

// Register a tokens provider for the language
monaco.languages.setMonarchTokensProvider('goink', {
  defaultToken: '',
  tokenPostfix: '.ink',
  tokenizer: {
    root: [[/(ink|goink|inky|inkle)/, ['inky']]]
  }
})

// Define a new theme that contains only rules that match this language
monaco.editor.defineTheme('goinkTheme', {
  base: 'vs',
  inherit: false,
  rules: [{ token: 'inky', foreground: '202020', fontStyle: 'bold italic' }]
})

const editor = monaco.editor.create(document.getElementById('editor'), {
  theme: 'goinkTheme',
  value: getCode(),
  language: 'goink'
})

function getCode () {
  return [
    '<goink ver 0.0.5-alpha>',
    "This is a go rewrite of inkle's ink - ",
    'a scripting language for writing interactive narrative.'
  ].join('\n')
}

// set cursor to the end of editor
editor.focus()

const range = editor.getModel().getFullModelRange()
editor.setPosition({
  lineNumber: range.endLineNumber,
  column: range.endColumn
})

var timeout = null
// send editor's content with delayed time
editor.onDidChangeModelContent(function () {
  if (timeout != null) clearTimeout(timeout)

  timeout = setTimeout(function () {
    fetch('http://localhost:9090/editor/onchange', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        value: editor.getValue()
      })
    })
      .then(function (resp) {
        return resp.json()
      })
      .then(function (json) {
        console.log('res', json)
      })
      .catch(function (error) {
        console.error('request failed', error)
      })
  }, 600)
})
