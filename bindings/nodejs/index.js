const ffi = require('ffi-napi');
const path = require('path');
const os = require('os');

// Detectar extensión de biblioteca según SO
const ext = os.platform() === 'win32' ? '.dll' : (os.platform() === 'darwin' ? '.dylib' : '.so');
const libPath = path.resolve(__dirname, './libplf' + ext);

// Cargar motor de PLF (Go)
let lib;
try {
    lib = ffi.Library(libPath, {
        'RenderPLF': ['string', ['string', 'string']]
    });
} catch (e) {
    console.warn('⚠️ PLF-JS: No se encontró la biblioteca nativa en ' + libPath);
}

/**
 * Renderiza un archivo PLF con variables desde Node.js.
 * @param {string} plfFile - Ruta al archivo .plf
 * @param {object} variables - Variables para el template
 * @returns {string} - Prompt renderizado
 */
function renderPlf(plfFile, variables) {
    if (!lib) throw new Error('Biblioteca nativa libplf no cargada.');
    
    const varsJson = JSON.stringify(variables);
    return lib.RenderPLF(plfFile, varsJson);
}

module.exports = {
    renderPlf
};
