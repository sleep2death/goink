import 'bulma/css/bulma.css'

import * as monaco from 'monaco-editor'

import Noty from 'noty'
import 'noty/lib/noty.css'
import 'noty/lib/themes/mint.css'

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
    root: [
      [/(ink|goink|inkle|inky)/, ['inky']],
      [/(end|End|END)/, ['inky']],
      [/\/\/*.*/, 'comment'],
      [/^\s*\*\**/, 'header'],
      [/^\s*--*/, 'header'],
      [/\s*->\s*/, 'header']
    ]
  }
})

monaco.languages.setLanguageConfiguration('goink', {
  comments: {
    lineComment: '//',
    blockComment: ['/*', '*/']
  },
  brackets: [
    ['{', '}'],
    ['[', ']'],
    ['(', ')']
  ],
  autoClosingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '`', close: '`', notIn: ['string'] },
    { open: '"', close: '"', notIn: ['string'] },
    { open: "'", close: "'", notIn: ['string', 'comment'] }
  ],
  surroundingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '`', close: '`' },
    { open: '"', close: '"' },
    { open: "'", close: "'" }
  ]
})

// Define a new theme that contains only rules that match this language
monaco.editor.defineTheme('goinkTheme', {
  base: 'vs',
  inherit: true,
  rules: [
    { token: 'inky', foreground: '202020', fontStyle: 'bold italic' },
    { token: 'header', foreground: '0366D6', fontStyle: 'bold' }
  ]
})

const editor = monaco.editor.create(document.getElementById('editor'), {
  theme: 'goinkTheme',
  value: getCode(),
  language: 'goink'
})

const model = editor.getModel()

function getCode () {
  return [
    '<goink ver 0.0.5-alpha>',
    "This is a go rewrite of inkle's ink - https://github.com/inkle/ink ",
    'a scripting language for writing interactive narrative.',
    "// Let's get started!",
    '* Here is a simple option.',
    '* Using "*" to add an option.',
    '- Gather line will be taken, when option runs out of content.',
    'Every story go to the end line finally. -> end'
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
const lnReg = /^(.+)\sln:\s(\d+)/
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
      .then((resp) => {
        return resp.json()
      })
      .then((json) => {
        if (json.error != null) {
          showError('parsing error...')

          const res = json.error.match(lnReg)
          if (res[2] != null) {
            const ln = parseInt(res[2])
            // console.log('some error happens @line:', ln)
            // set error marker
            monaco.editor.setModelMarkers(model, '', [
              {
                severity: monaco.MarkerSeverity.Error,
                message: res[1],
                startColumn: 0,
                startLineNumber: ln,
                endColumn: 1000,
                endLineNumber: ln
              }
            ])
          }
        } else {
          // clear markers
          monaco.editor.setModelMarkers(model, '', [])
        }
      })
      .catch(function () {
        showError('can not fetch from server...')
      })
  }, 600)
})

function showError (error) {
  new Noty({
    type: 'error',
    theme: 'mint',
    timeout: 1500,
    layout: 'topRight',
    progressBar: false,
    text: error
  }).show()
}
