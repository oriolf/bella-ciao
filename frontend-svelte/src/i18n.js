import { addMessages, init, getLocaleFromNavigator } from 'svelte-i18n'
import en from './lang/en.js'
import ca from './lang/ca.js'
import es from './lang/es.js'

function setupI18n() {
    addMessages('en', en)
    addMessages('ca', ca)
    addMessages('es', es)

    init({
        fallbackLocale: 'en',
        initialLocale: getLocaleFromNavigator(),
    })
}

export { setupI18n };