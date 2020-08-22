require.config({ paths: { vs: 'static/js/monaco/vs' } })
require(['vs/editor/editor.main'], function () {
  // Custom goink language specific
  monaco.languages.register({ id: 'goink' })

  // Register a tokens provider for the language
  monaco.languages.setMonarchTokensProvider('goink', {
    defaultToken: '',
    // escape codes
    control: /[\\`*_[\]{}()#+\-.!]/,
    noncontrol: /[^\\`*_[\]{}()#+\-.!]/,
    escapes: /\\(?:@control)/,

    tokenizer: {
      root: [
        [
          /^(\s*)(=+)((?:[^\\=]|@escapes)+)((?:=+)?)/,
          ['white', 'knot-header', 'knot', 'knot-header']
        ],
        [/goink/, 'goink'],
        [/\/\/.+/, 'comment'],
        [/#(\s*\w+)+/, 'tag']
      ]
    }
  })

  // Register theme of ink
  monaco.editor.defineTheme('inkTheme', {
    base: 'vs',
    inherit: true,
    rules: [
      { token: 'goink', foreground: '202020', fontStyle: 'bold italic' },
      { token: 'knot-header', foreground: '808080', fontStyle: 'bold' },
      { token: 'knot', foreground: '202020', fontStyle: 'bold' },
      { token: 'comment', foreground: '808080', fontStyle: 'italic' },
      { token: 'tag', fontStyle: 'bold' }
    ]
  })

  monaco.editor.create(document.getElementById('container'), {
    value: getCode(),
    theme: 'inkTheme',
    language: 'goink',
    minimap: {
      enabled: false
    }
  })
})

function getCode () {
  const input = [
    'Hello, world!',
    'This is "goink", a scripting language for writing interactive narrative.',
    '',
    '// Here is a single line comment.',
    'Multiline comments is NOT implemented yet.',
    '',
    'ink provides a simple system for tagging lines of content, with hashtags.',
    'A line of normal game-text. # colour it blue # or red // or both',
    '',

    'Input is offered to the player via text choices. A text choice is indicated by an * character.',
    'If no other flow instructions are given, once made, the choice will flow into the next line of text.',
    '* Hello back!',
    '\tNice to hear from you!'
  ]
  return input.join('\n')
}
