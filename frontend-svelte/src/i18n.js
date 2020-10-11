import { addMessages, init, getLocaleFromNavigator } from 'svelte-i18n'
import en from './en.js'
import ca from './ca.js'

function setupI18n() {
    addMessages('en', en)
    addMessages('ca', ca)

    init({
        fallbackLocale: 'en',
        initialLocale: getLocaleFromNavigator(),
    })
}

export { setupI18n };