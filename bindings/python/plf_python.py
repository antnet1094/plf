import ctypes
import json
import os
import sys

def test_bridge():
    lib_path = "./libplf.dll" if sys.platform == "win32" else "./libplf.so"
    
    if not os.path.exists(lib_path):
        print(f"ERROR: {lib_path} no encontrado.")
        return

    try:
        lib = ctypes.CDLL(lib_path)
        
        # Configurar tipos de entrada/salida para RenderPLF
        lib.RenderPLF.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
        lib.RenderPLF.restype = ctypes.c_char_p

        print("--- Iniciando prueba PLF Python Bridge ---")
        
        plf_file = b"examples/sysadmin.plf"
        variables = json.dumps({"mensaje_usuario": "hola desde python"}).encode('utf-8')
        
        # Llamar a la función de Go
        result_ptr = lib.RenderPLF(plf_file, variables)
        
        if result_ptr:
            result = result_ptr.decode('utf-8')
            print("\nResultado del motor PLF (Go) consumido desde Python:\n")
            print(result)
        else:
            print("ERROR: Recibido puntero nulo desde Go.")
            
    except Exception as e:
        print(f"ERROR al cargar o ejecutar el bridge: {e}")

if __name__ == "__main__":
    test_bridge()
