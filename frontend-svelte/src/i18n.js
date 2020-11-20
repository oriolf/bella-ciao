import { addMessages, init, getLocaleFromNavigator } from 'svelte-i18n'
import en from './lang/en.js'
import ca from './lang/ca.js'
import es from './lang/es.js'

function getInitialLocale(storage) {
    if (storage) {
        const lang = storage.getItem('bella-ciao.lang');
        if (lang) {
            return lang;
        }
    }
    return getLocaleFromNavigator();
}

function setupI18n(storage) {
    addMessages('en', en)
    addMessages('ca', ca)
    addMessages('es', es)

    init({
        fallbackLocale: 'en',
        initialLocale: getInitialLocale(storage),
    })
}

export { setupI18n };