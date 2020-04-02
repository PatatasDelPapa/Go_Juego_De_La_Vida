package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
)

type world struct {
	MAPA [][]bool
	NRO  int
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

	var chans [124]chan [][]bool
	for i := range chans {
		chans[i] = make(chan [][]bool)
	}

	resultado := make(chan world, 32)
	var wg sync.WaitGroup

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
	print("---------------\n")
	mapa = rellenar(mapa, semilla)
	newWorld := world{MAPA: mapa, NRO: 0}
	renderizar(newWorld.MAPA)

	// al final de cada generacion se realiza un wg.Wait() y luego se reorganiza y renderiza el estado actual del mapa
	// 	for i := 0; i < generaciones; i++ {
	// 		for j := 0; j < nroGorrutinas; j++ {
	// 			mapaGorrutina := calcularMapa(mapa, nroGorrutinas, filas, columnas, j)
	// 			wg.Add(1)
	// 			if i == 0 {
	// 				go procesar(mapaGorrutina, &wg, true, false, j, nroGorrutinas, filas, chans)
	// 			} else if i == nroGorrutinas-1 {
	// 				go procesar(mapaGorrutina, &wg, false, true, j, nroGorrutinas, filas, chans)
	// 			} else {
	// 				go procesar(mapaGorrutina, &wg, false, false, j, nroGorrutinas, filas, chans)
	// 			}
	// 		}
	// 		wg.Wait()
	// 		// REORGANIZAR TODOS LOS MAPAS Y LUEGO RENDERIZAR
	// 		// reorganizar(mapas de cada gorrutina)
	// 		renderizar(mapa)
	// 	}

	_ = wg
	_ = resultado
	wg.Add(1)
	mundoGorrutina := calcularMapa(mapa, nroGorrutinas, filas, columnas, 0)
	println("------------------------------------------")
	renderizar(mundoGorrutina.MAPA)
	println("------------------------------------------")
	go procesar(mundoGorrutina, &wg, true, false, 0, nroGorrutinas, filas, chans, resultado)
	wg.Wait()
	mundito := <-resultado
	renderizar(mundito.MAPA)

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
func rellenar(mapa [][]bool, semilla int) [][]bool {
	s := rand.NewSource(42)
	r := rand.New(s)
	// max := ((area.fin.x - area.inicio.y) * (area.fin.y - area.inicio.y))
	max := (len(mapa) * len(mapa[0]))
	if semilla > max {
		semilla = max
	}
	for i := 0; i < semilla; i++ {
		x := r.Intn(len(mapa))
		y := (r.Intn(len(mapa[0])))
		if mapa[x][y] {
			i--
		} else {
			mapa[x][y] = true
		}
	}
	return mapa
}

// FUNCION QUE BUSCA RETORNAR UN NUEVO MAPA DE DIMENSIONES [filas][((k+1)*bloque)]
// QUE COPIE EL MAPA ORIGINAL DESDE [0][(k*bloque)] hasta [filas-1][((k+1)*bloque-1)]
func calcularMapa(mapa [][]bool, hilos int, filas, columnas, k int) world {

	bloque := columnas / hilos

	resto := columnas % hilos

	if resto != 0 {
		panic("Los bloques deben ser de igual tamaño, el tamaño, es decir, el modulo de la cantidad de columnas por la cantidad de rutinas debe ser igual a 0")
	}

	// COPIAR A ESTE NUEVO MAPA DESDE EL MAPA ORIGINAL DESDE [0][(i*bloque)] hasta [filas-1][((k+1)*bloque-1)]
	// RETORNAR EL NUEVO MAPA COPIADO
	columnaMin := k * bloque
	columnaMax := (k + 1) * bloque

	newMapa := make([][]bool, len(mapa))
	for i := range mapa {
		newMapa[i] = make([]bool, (columnaMax - columnaMin))
		copy(newMapa[i], mapa[i])
	}

	// println("------------------------------------------")
	// fmt.Println("Num Bloques = ", bloque)
	// fmt.Println("k * bloque = ", columnaMin, "\n",
	// 	"(k+1)*bloque = ", columnaMax)
	// println("------------------------------------------")
	// renderizar(newMapa)
	// println("------------------------------------------")

	newWorld := world{MAPA: newMapa, NRO: 0}

	return newWorld
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
func procesar(mundo world, wg *sync.WaitGroup, inicio, fin bool, k, n, filas int, chans [124]chan [][]bool, resultado chan world) {

	// nota: "k" es el numero actual de la gorrutina el cual va desde k = 0 hasta k = (numero total de gorrutinas - 1)
	// el numero actual de la gorrutina  es util para el thread principal que se encargara de reorganizar el mapa completo en base a los sub mapas de
	// las gorrutinas

	// fmt.Println(
	// 	"K: ", k, "\n",
	// 	"N: ", n, "\n",
	// 	"Filas: ", filas)

	defer wg.Done()
	// mapa [][]bool, inicio, fin bool, k, n, filas, columnas int, chans []chan []bool
	nuevoEstado(mundo, inicio, fin, k, n, filas, chans, resultado)

}

// FUNCION QUE REVISARA LAS CONDICIONES DE BORDE DE LA GORRUTINA Y UTILIZARA CHANNELS PARA OBTENER Y ENVIAR LOS BORDES NECESARIOS
// REALIZARA UNA EXTENSION FANTASMA DEL AREA QUE TIENE
// LLAMARA A LA FUNCION QUE SE ENCARGUE DE ACTUALIZAR EL ESTADO ACTUAL DE LA CELDA PARA CADA CELDA QUE TENGA
// RETORNARA EL NUEVO ESTADO DE SU AREA
func nuevoEstado(mundo world, inicio, fin bool, k, n, filas int, chans [124]chan [][]bool, resultado chan world) {
	if n == 1 {
		// CASO PARA CUANDO ES UNA GORRUTINA
		resultado <- mundo
	}
	if inicio {
		println(" INICIO TRUE")
		entrada := chans[0]
		salida := chans[1]
		// borde := len(mundo.MAPA[0])
		// -----------------------------------------//
		// bordeDerecho := make([][]bool, len(mundo.MAPA))
		// for i := range mundo.MAPA {
		// 	bordeDerecho[i] = make([]bool, 1)
		// 	copy(bordeDerecho[i], mundo.MAPA[i])
		// }
		// -----------------------------------------//

		_ = entrada
		_ = salida

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range viejoMapa {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		// USANDO NEW MAPA HACER LOS CALCULOS DEL NUEVO ESTADO

		mundo.MAPA = newMapa
		resultado <- mundo

	} else if fin {
		println(" FIN TRUE ")
		entrada := chans[n-2]
		salida := chans[n-1]
		// -----------------------------------------//
		// bordeDerecho := mundo.MAPA[0:filas][0]
		bordeIzquierdo := make([][]bool, len(mundo.MAPA))
		for i := range mundo.MAPA {
			bordeIzquierdo[i] = make([]bool, 1)
			copy(bordeIzquierdo[i], mundo.MAPA[i])
		}
		// -----------------------------------------//

		_ = entrada
		_ = salida

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range mundo.MAPA {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		mundo.MAPA = newMapa
		resultado <- mundo
	} else {
		println(" NO INICIO NO FIN TRUE ")
		entradaIzquierda := chans[k*4-2]
		salidaIzquierda := chans[k*4-1]
		entradaDerecha := chans[k*4]
		salidaDerecha := chans[k*4+1]

		eBordeIzquierdo := <-entradaIzquierda
		eBordeDerecho := <-entradaDerecha
		_ = entradaDerecha
		_ = salidaDerecha
		_ = entradaIzquierda
		_ = salidaIzquierda
		_ = eBordeIzquierdo
		_ = eBordeDerecho

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range mundo.MAPA {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		mundo.MAPA = newMapa
		resultado <- mundo
	}
}

// CALCULO DE LOS NUEVOS ESTADOS: (SE REALIZA DENTRO DEL PROPIO IF)
// Extension fantasma de su mapa
// Recorrer el mapa, contar cuantos vecinos vivos tiene cada celda
// Mandar esa informacion a la funcion que calcula si la celda vive o muere
// Actualizar la informacion en un nuevo mapa temporal para no arruinar el calculo de las demas celdas
// Retornar el mapa temporal con la informacion de las celdas actualizadas
// }
