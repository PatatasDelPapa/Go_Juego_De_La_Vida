package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
)

type celda struct {
	x int
	y int
}

type tablero struct {
	inicio celda
	fin    celda
}

func main() {
	//  -ng NUM_GORUTINAS -r NUM_FILAS -c NUM_COLS -i GENERACIONES -s SEMILLA
	args := os.Args
	generaciones := 1
	columnas := 1
	filas := 1
	semilla := 1
	nroGorrutinas := 1

	for i, arg := range args {
		switch arg {
		case "-ng":
			nroGorrutinas, _ = strconv.Atoi(args[i+1])
		case "-r":
			filas, _ = strconv.Atoi(args[i+1])
		case "-c":
			columnas, _ = strconv.Atoi(args[i+1])
		case "-i":
			generaciones, _ = strconv.Atoi(args[i+1])
		case "-s":
			semilla, _ = strconv.Atoi(args[i+1])
		}
	}

	var chans [124]chan []bool
	for i := range chans {
		chans[i] = make(chan []bool)
	}

	//  -ng NUM_GORUTINAS -r NUM_FILAS -c NUM_COLS -i GENERACIONES -s SEMILLA
	fmt.Println(" Num Gorrutinas = ", nroGorrutinas, "\n",
		"Num Filas = ", filas, "\n",
		"Num Columnas = ", columnas, "\n",
		"Generaciones = ", generaciones, "\n",
		"Semilla = ", semilla)

	mapa := make([][]bool, filas)
	for i := 0; i < len(mapa); i++ {
		mapa[i] = make([]bool, columnas)
	}
	renderizar(mapa)
	print("---------------\n")
	rellenar(mapa, semilla, tablero{inicio: celda{x: 0, y: 0}, fin: celda{x: filas, y: columnas}})
	renderizar(mapa)
	var wg sync.WaitGroup

	// al final de cada generacion se realiza un wg.Wait() y luego se reorganiza y renderiza el estado actual del mapa
	for i := 0; i < generaciones; i++ {
		for j := 0; j < nroGorrutinas; j++ {
			// area := calcularTablero(nroGorrutinas, filas, columnas)
			wg.Add(1)
			// if i == 0 {
			// 	go procesar(area[i], &wg, true, false)
			// } else if i == nroGorrutinas-1 {
			// 	go procesar(area[i], &wg, false, true)
			// } else {
			// 	go procesar(area[i], &wg, false, false)
			// }

		}
		wg.Wait()
		// REORGANIZAR TODOS LOS MAPAS Y LUEGO RENDERIZAR
		// reorganizar(mapas de cada gorrutina)
		renderizar(mapa)
	}
}

// FUNCION QUE IMPRIME EN PANTALLA EL RESULTADO DE LA ITERACION ACTUAL DEL ESTADO DEL MAPA
func renderizar(mapa [][]bool) {
	for i := range mapa {
		for _, v := range mapa[i] {
			if v {
				print("■ ")
			} else {
				print("□ ")
			}
		}
		print("\n")
	}
	// print("\n")
}

// FUNCION QUE RELLENA UN AREA CON TANTAS SEMILLAS SE SOLICITEN O HASTA QUE SE LLENE TODA EL AREA
func rellenar(mapa [][]bool, semilla int, area tablero) {
	s := rand.NewSource(42)
	r := rand.New(s)
	max := ((area.fin.x - area.inicio.y) * (area.fin.y - area.inicio.y))
	if semilla > max {
		semilla = max
	}
	for i := 0; i < semilla; i++ {
		x := r.Intn(area.fin.x)
		y := (r.Intn(area.fin.y-area.inicio.y) + area.inicio.y)
		if mapa[x][y] {
			i--
		} else {
			mapa[x][y] = true
		}
	}
}

// FUNCION QUE BUSCA RETORNAR UN NUEVO MAPA DE DIMENSIONES [filas][((k+1)*bloque)]
// QUE COPIE EL MAPA ORIGINAL DESDE [0][(k*bloque)] hasta [filas-1][((k+1)*bloque-1)]
func calcularMapa(mapa [][]bool, hilos int, filas, columnas, k int) [][]bool {

	bloque := columnas / hilos
	resto := columnas % hilos

	if resto != 0 {
		panic("Los bloques deben ser de igual tamaño, el tamaño, es decir, el modulo de la cantidad de columnas por la cantidad de rutinas debe ser igual a 0")
	}

	newMapa := make([][]bool, filas)
	for i := 0; i < len(newMapa); i++ {
		newMapa[i] = make([]bool, ((k + 1) * bloque))
	}

	// COPIAR A ESTE NUEVO MAPA DESDE EL MAPA ORIGINAL DESDE [0][(i*bloque)] hasta [filas-1][((k+1)*bloque-1)]
	// RETORNAR EL NUEVO MAPA COPIADO

	newMapa = mapa[0:filas][(k * bloque) : (k+1)*bloque]

	return newMapa
}

// FUNCION QUE SE ENCARGA DE EVALUAR SI LA CELDA CONTINUA VIVA O NO
func transiciones(celda bool, con int) bool {
	if celda {
		switch {
		case con < 3:
			return false
		case con == 3 || con == 4:
			return true
		case con > 4:
			return false
		default:
			panic("Esta linea no deberia pasar nunca")
		}
	} else {
		if con == 3 {
			return true
		}
	}
	return false
}

// FUNCION QUE SE EJECUTA COMO GORRUTINA
// SE LE ENTREGA SU SUB-MAPA, EL WAITGROUP PARA SINCRONIZAR, DOS BOOLEANOS PARA INDICAR SI ES INICIO O FINAL Y EL NUMERO DE GORRUTINA QUE ES
// SE ENCARGARA DE LLAMAR A TODAS LAS FUNCIONES QUE REALIZAN OPERACIONES PARA EVALUAR EL PROXIMO ESTADO DE SU SUB-MAPA
// AL TERMINAR DEVOLVERA EL NUEVO ESTADO DE SU SUB-MAPA AL THREAD PRINCIPAL Y SU NUMERO DE GORRUTINA
func procesar(mapa [][]bool, wg *sync.WaitGroup, inicio, fin bool, k, n, filas, columnas int, chans []chan bool) ([][]bool, int) {

	// nota: "k" es el numero actual de la gorrutina el cual va desde k = 0 hasta k = (numero total de gorrutinas - 1) siendo 0 la gorrutina con la columna inicial
	//       y (numero total de gorrutinas - 1) la gorrutina con la columna final las cuales tienen condiciones especiales de borde que las demas gorrutinas.
	// el numero actual de la gorrutina tambien es util para el thread principal que se encargara de reorganizar el mapa completo en base a los sub mapas de
	// las gorrutinas

	defer wg.Done()

	// newMapa = nuevoEstado(mapa, inicio, fin, k, n, filas, columnas)
	// return newMapa, k

	return mapa, k
}

// FUNCION QUE REVISARA LAS CONDICIONES DE BORDE DE LA GORRUTINA Y UTILIZARA CHANNELS PARA OBTENER Y ENVIAR LOS BORDES NECESARIOS
// REALIZARA UNA EXTENSION FANTASMA DEL AREA QUE TIENE
// LLAMARA A LA FUNCION QUE SE ENCARGUE DE ACTUALIZAR EL ESTADO ACTUAL DE LA CELDA PARA CADA CELDA QUE TENGA
// RETORNARA EL NUEVO ESTADO DE SU AREA
func nuevoEstado(mapa [][]bool, inicio, fin bool, k, n, filas, columnas int, chans []chan []bool) [][]bool {
	if inicio {
		entrada := chans[0]
		salida := chans[1]
		borde := len(mapa[0])
		bordeIzquierdo := mapa[0:filas][borde]
		salida <- bordeIzquierdo
		bordeDerecho := <-salida
		_ = entrada
		_ = salida
		_ = bordeIzquierdo
		_ = bordeDerecho
		// La gorrutina tiene la columna inicial
		// Enviar su borde a su unico vecino
		// Buscar borde de su unico vecino
		// Realizar los pasos que aparecen en los comentarios al final
		return mapa
	} else if fin {
		entrada := chans[n-2]
		salida := chans[n-1]
		bordeDerecho := mapa[0:filas][0]
		salida <- bordeDerecho
		bordeIzquierdo := <-entrada
		_ = entrada
		_ = salida
		_ = bordeDerecho
		_ = bordeIzquierdo
		// La gorrutina tiene la columna final
		// Enviar su borde a su unico vecino
		// Buscar borde de su unico vecino
		// Realizar los pasos que aparecen en los comentarios al final
		return mapa
	} else {
		entradaIzquierda := chans[k*4-2]
		salidaIzquierda := chans[k*4-1]
		entradaDerecha := chans[k*4]
		salidaDerecha := chans[k*4+1]
		borde := len(mapa[0])
		sBordeIzquierdo := mapa[0:filas][0]
		sBordeDerecho := mapa[0:filas][borde]
		salidaIzquierda <- sBordeIzquierdo
		salidaDerecha <- sBordeDerecho
		eBordeIzquierdo := <-entradaIzquierda
		eBordeDerecho := <-entradaDerecha
		_ = entradaDerecha
		_ = salidaDerecha
		_ = entradaIzquierda
		_ = salidaIzquierda
		_ = eBordeIzquierdo
		_ = eBordeDerecho
		// La gorrutina tiene una columna intermedia
		// Enviar su borde respectivo al vecino respectivo
		// Buscar borde a sus dos vecinos
		// Realizar los pasos que aparecen en los comentarios al final
		return mapa
	}
	// CALCULO DE LOS NUEVOS ESTADOS: (SE REALIZA DENTRO DEL PROPIO IF)
	// Extension fantasma de su mapa
	// Recorrer el mapa, contar cuantos vecinos vivos tiene cada celda
	// Mandar esa informacion a la funcion que calcula si la celda vive o muere
	// Actualizar la informacion en un nuevo mapa temporal para no arruinar el calculo de las demas celdas
	// Retornar el mapa temporal con la informacion de las celdas actualizadas
}
