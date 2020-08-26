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
    "This is a go rewrite of inkle's ink // see: https://github.com/inkle/ink ",
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
// send editor's content with delayed time
editor.onDidChangeModelContent(function () {
  if (timeout != null) clearTimeout(timeout)
  timeout = setTimeout(function () {
    onChange()
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

function addErrorMarkers (model, errs) {
  var markers = []
  errs.forEach((e) => {
    if (e.ln > 0) {
      markers.push({
        severity: monaco.MarkerSeverity.Error,
        message: e.msg,
        startColumn: 0,
        startLineNumber: e.ln,
        endColumn: model.getLineMaxColumn(e.ln),
        endLineNumber: e.ln
      })
    } else {
      showError(e.msg)
    }
  })

  monaco.editor.setModelMarkers(model, '', markers)
}

function onChange () {
  fetch('http://localhost:9090/editor/onchange', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      value: editor.getValue(),
      uuid: uuid
    })
  })
    .then((resp) => {
      return resp.json()
    })
    .then((json) => {
      // if error is a line marker
      if (json.errors != null) {
        showError('story parsing error')
        addErrorMarkers(model, json.errors)
      } else {
        // set uuid from server
        if (uuid === '') {
          uuid = json.uuid
        } else if (uuid !== json.uuid) {
          throw new Error('conflict user id from server')
        }
        //
        // clear markers
        monaco.editor.setModelMarkers(model, '', [])
        if (json.section != null) {
          const content = document.getElementById('content')
          content.innerHTML = ''

          const c = document.createElement('div')
          c.innerText = json.section.text
          content.appendChild(c)

          const options = document.getElementById('options')

          options.innerHTML = ''
          if (json.section.opts) {
            // add separator
            const sep = document.createElement('div')
            sep.className = 'separator'
            sep.innerText = 'Options'
            content.appendChild(sep)

            json.section.opts.forEach((opt, idx) => {
              options.innerHTML += `<li><a href="#" onclick="choose(${idx})">${opt}</a></li>`
            })
          }
        }
      }
    })
    .catch(function (e) {
      showError('Oops, ' + e.toString())
    })
}

var uuid = ''

window.choose = (idx) => {
  fetch('http://localhost:9090/review/onchoose', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      uuid: uuid,
      index: parseInt(idx)
    })
  })
    .then((resp) => {
      return resp.json()
    })
    .then((json) => {
      // if error is a line marker
      if (json.errors != null) {
        showError('story parsing error')
        addErrorMarkers(model, json.errors)
      } else {
        // set uuid from server
        if (uuid !== json.uuid) {
          throw new Error('conflict user id from server')
        }
        // clear markers
        monaco.editor.setModelMarkers(model, '', [])

        if (json.section != null) {
          const content = document.getElementById('content')

          const seps = content.getElementsByClassName('separator')

          for (let i = 0; i < seps.length; i++) {
            seps[i].innerText = '-'
          }

          const c = document.createElement('div')
          c.innerText = json.section.text
          content.appendChild(c)

          const options = document.getElementById('options')
          options.innerHTML = ''

          if (json.section.opts) {
            // add separator
            const sep = document.createElement('div')
            sep.className = 'separator'
            sep.innerText = 'Options'
            content.appendChild(sep)

            json.section.opts.forEach((opt, idx) => {
              options.innerHTML += `<li><a href="#" onclick="choose(${idx})">${opt}</a></li>`
            })
          } else if (json.section.end) {
            const end = document.getElementById('end')
            end.classList.remove('hidden')
            /* const sep = document.createElement('div')
            sep.className = 'separator'
            sep.innerText = 'THE END'
            options.parentElement.appendChild(sep) */
          }
        }
      }
    })
    .catch(function (e) {
      showError('Oops, ' + e.toString())
    })
}

onChange()
