import { globalIgnores } from 'eslint/config'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'
import pluginOxlint from 'eslint-plugin-oxlint'
import skipFormatting from 'eslint-config-prettier/flat'

// To allow more languages other than `ts` in `.vue` files, uncomment the following lines:
// import { configureVueProject } from '@vue/eslint-config-typescript'
// configureVueProject({ scriptLangs: ['ts', 'tsx'] })
// More info at https://github.com/vuejs/eslint-config-typescript/#advanced-setup

export default defineConfigWithVueTs(
  {
    name: 'app/files-to-lint',
    files: ['**/*.{vue,ts,mts,tsx}'],
  },

  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**']),

  ...pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,

  ...pluginOxlint.buildFromOxlintConfigFile('.oxlintrc.json'),

  {
    name: 'app/component-naming',
    rules: {
      'vue/component-api-style': ['error', ['options']],
      'vue/multi-word-component-names': 'off',
      'vue/no-reserved-component-names': 'off',
      'vue/no-restricted-syntax': [
        'error',
        {
          selector: 'VAttribute[directive=false][key.name=/^aria-/]',
          message: 'ARIA attributes are not used in this project.',
        },
        {
          selector: 'VAttribute[directive=true][key.argument.name=/^aria-/]',
          message: 'ARIA attribute bindings are not used in this project.',
        },
        {
          selector: 'Property[key.value=/^aria-/]',
          message: 'Do not pass ARIA attributes through attribute objects.',
        },
      ],
    },
  },

  skipFormatting,
)
