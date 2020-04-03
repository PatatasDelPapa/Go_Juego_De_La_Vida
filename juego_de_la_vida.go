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
		chans[i] = make(chan [][]bool, 100)
	}

	resultado := make(chan world)
	var wg, jo sync.WaitGroup

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
	_ = newWorld
	renderizar(newWorld.MAPA)
	print("---------------\n")

	// al final de cada generacion se realiza un wg.Wait() y luego se reorganiza y renderiza el estado actual del mapa
	for i := 0; i < generaciones; i++ {
		jo.Add(nroGorrutinas)
		wg.Add(nroGorrutinas)
		for j := 0; j < nroGorrutinas; j++ {
			mundoGorrutina := calcularMapa(mapa, nroGorrutinas, filas, columnas, j)
			// renderizar(mundoGorrutina.MAPA)
			// print("---------------\n")
			if j == 0 {
				go procesar(mundoGorrutina, &wg, &jo, true, false, j, nroGorrutinas, filas, chans, resultado)
			} else if j == nroGorrutinas-1 {
				go procesar(mundoGorrutina, &wg, &jo, false, true, j, nroGorrutinas, filas, chans, resultado)
			} else {
				go procesar(mundoGorrutina, &wg, &jo, false, false, j, nroGorrutinas, filas, chans, resultado)
			}
		}
		for j := 0; j < nroGorrutinas; j++ {
			// fmt.Println("J = ", j, "Antes de <- resultado")
			<-resultado
			// fmt.Println("Despues de <- Resultado")
		}
		// println("Antes de wg.Wait")
		wg.Wait()
		// println("Despues de wg.Wait")
		// REORGANIZAR TODOS LOS MAPAS Y LUEGO RENDERIZAR
		// reorganizar(mapas de cada gorrutina)
		println("------------------------NUEVA GENERACION---------------------------------------")
		// renderizar(mapa)
	}

	// wg.Add(1)
	// mundoGorrutina := calcularMapa(mapa, nroGorrutinas, filas, columnas, 0)
	// go procesar(mundoGorrutina, &wg, true, false, 0, nroGorrutinas, filas, chans, resultado)
	// wg.Wait()
	// mundito := <-resultado
	// renderizar(mundito.MAPA)

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
		panic("Los bloques deben ser de igual tamaño, es decir, el modulo de la cantidad de columnas por la cantidad de rutinas debe ser igual a 0")
	}

	// COPIAR A ESTE NUEVO MAPA DESDE EL MAPA ORIGINAL DESDE [0][(i*bloque)] hasta [filas-1][((k+1)*bloque-1)]
	// RETORNAR EL NUEVO MAPA COPIADO
	columnaMin := k * bloque
	columnaMax := (k + 1) * bloque

	newMapa := make([][]bool, len(mapa))
	// for i := range mapa {
	// 	newMapa[i] = make([]bool, (columnaMax - columnaMin))
	// 	copy(newMapa[i], mapa[i])
	// }

	for i := range mapa {
		newMapa[i] = make([]bool, (columnaMax - columnaMin))
		for j := 0; j < len(newMapa[i]); j++ {
			newMapa[i][j] = mapa[i][columnaMin+j]
		}
	}

	// println("------------------------------------------")
	// fmt.Println("Num Bloques = ", bloque)
	// fmt.Println("k * bloque = ", columnaMin, "\n",
	// 	"(k+1)*bloque = ", columnaMax)
	// println("------------------------------------------")
	// renderizar(newMapa)
	// println("------------------------------------------")

	newWorld := world{MAPA: newMapa, NRO: k}

	return newWorld
}

// FUNCION QUE SE ENCARGA DE EVALUAR SI LA CELDA CONTINUA VIVA O NO
func transiciones(celda bool, con int) bool {
	if celda {
		switch {
		case con < 3:
			fmt.Print(" Celda Viva Muere Por Pocos Vecinos\n")
			return false
		case con == 3 || con == 4:
			fmt.Print(" Celda Viva Sobrevive\n")
			return true
		case con > 4:
			fmt.Print(" Celda Viva Muere Por Sobrepoblamiento\n")
			return false
		default:
			panic("Esta linea no deberia pasar nunca")
		}
	} else {
		if con == 3 {
			fmt.Print(" Celda Muerta Revive\n")
			return true
		}
	}
	fmt.Print(" Celda Muerta Permanece Muerta\n")
	return false
}

// FUNCION QUE SE EJECUTA COMO GORRUTINA
// SE LE ENTREGA SU SUB-MAPA, EL WAITGROUP PARA SINCRONIZAR, DOS BOOLEANOS PARA INDICAR SI ES INICIO O FINAL Y EL NUMERO DE GORRUTINA QUE ES
// SE ENCARGARA DE LLAMAR A TODAS LAS FUNCIONES QUE REALIZAN OPERACIONES PARA EVALUAR EL PROXIMO ESTADO DE SU SUB-MAPA
// AL TERMINAR DEVOLVERA EL NUEVO ESTADO DE SU SUB-MAPA AL THREAD PRINCIPAL Y SU NUMERO DE GORRUTINA
func procesar(mundo world, wg, jo *sync.WaitGroup, inicio, fin bool, k, n, filas int, chans [124]chan [][]bool, resultado chan world) {
	// nota: "k" es el numero actual de la gorrutina el cual va desde k = 0 hasta k = (numero total de gorrutinas - 1)
	// el numero actual de la gorrutina  es util para el thread principal que se encargara de reorganizar el mapa completo en base a los sub mapas de
	// las gorrutinas

	// fmt.Println(
	// 	"K: ", k, "\n",
	// 	"N: ", n, "\n",
	// 	"Filas: ", filas)
	// println("-----------------------------")
	// fmt.Println("Inicio: ", inicio)
	// fmt.Println("Fin: ", fin)
	// fmt.Println("Nro Gorrutina: ", k)
	// println("-----------------------------")
	// mapa [][]bool, inicio, fin bool, k, n, filas, columnas int, chans []chan []bool
	nuevoEstado(mundo, inicio, fin, k, n, filas, chans, resultado, wg, jo)

}

// FUNCION QUE REVISARA LAS CONDICIONES DE BORDE DE LA GORRUTINA Y UTILIZARA CHANNELS PARA OBTENER Y ENVIAR LOS BORDES NECESARIOS
// REALIZARA UNA EXTENSION FANTASMA DEL AREA QUE TIENE
// LLAMARA A LA FUNCION QUE SE ENCARGUE DE ACTUALIZAR EL ESTADO ACTUAL DE LA CELDA PARA CADA CELDA QUE TENGA
// RETORNARA EL NUEVO ESTADO DE SU AREA
func nuevoEstado(mundo world, inicio, fin bool, k, n, filas int, chans [124]chan [][]bool, resultado chan world, wg, jo *sync.WaitGroup) {

	if n == 1 {
		// CASO PARA CUANDO ES UNA GORRUTINA
		resultado <- mundo
	}
	if inicio {
		entrada := chans[0]
		salida := chans[1]
		_ = entrada
		_ = salida

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range viejoMapa {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		sBordeDerecho := make([][]bool, len(newMapa))
		for i := range sBordeDerecho {
			sBordeDerecho[i] = make([]bool, 1)
		}

		for i := 0; i < len(newMapa); i++ {
			for j := 0; j < 1; j++ {
				sBordeDerecho[i][j] = newMapa[i][len(newMapa[i])-1]
			}
		}

		jo.Done()
		// fmt.Println("Gorrutina: ", k, " Before jo.Wait()")
		jo.Wait()

		// fmt.Println("Gorrutina: ", k, " Mandando sBordeDerecho")
		salida <- sBordeDerecho
		// fmt.Println("Gorrutina: ", k, " Envio sBordeDerecho")

		// fmt.Println("Gorrutina: ", k, " Esperando eBordeDerecho")
		eBordeDerecho := <-entrada
		// fmt.Println("Gorrutina: ", k, " Recibio eBordeDerecho")

		_ = eBordeDerecho

		mapaExtendido := make([][]bool, len(newMapa))
		for i := range newMapa {
			mapaExtendido[i] = make([]bool, len(newMapa[i]))
			copy(mapaExtendido[i], newMapa[i])
		}
		// println("------------------------")
		// renderizar(mapaExtendido)

		for i := 0; i < len(eBordeDerecho); i++ {
			for j := 0; j < len(eBordeDerecho[i]); j++ {
				mapaExtendido[i] = append(mapaExtendido[i], eBordeDerecho[i][j])
			}
		}

		println("-----------ANTES--------------")
		renderizar(newMapa)
		println()

		for i := 0; i < len(mapaExtendido); i++ {
			for j := 0; j < len(mapaExtendido[i])-1; j++ {
				newMapa[i][j] = vecinos(mapaExtendido, i, j)
			}
		}
		println()
		println("----------DESPUES------------")
		renderizar(newMapa)

		// USANDO NEW MAPA HACER LOS CALCULOS DEL NUEVO ESTADO

		for i := range newMapa {
			for j := range newMapa[i] {
				newMapa[i][j] = mapaExtendido[i][j]
			}
		}

		// renderizar(newMapa)

		mundo.MAPA = newMapa
		// fmt.Println("Gorrutina: ", k, " Mandando mensaje por el channel resultado")
		resultado <- mundo
		// fmt.Println("Gorrutina: ", k, " Mensaje enviado por el channel resultado")
		wg.Done()

	} else if fin {
		var entrada chan [][]bool
		var salida chan [][]bool
		salida = chans[k*2-2]
		entrada = chans[k*2-1]

		_, _ = salida, entrada
		// println(" FIN TRUE ")

		sBordeIzquierdo := make([][]bool, len(mundo.MAPA))
		for i := range sBordeIzquierdo {
			sBordeIzquierdo[i] = make([]bool, 1)
			sBordeIzquierdo[i][0] = mundo.MAPA[i][0]
		}

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range mundo.MAPA {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		jo.Done()
		// fmt.Println("Gorrutina: ", k, " Before jo.Wait()")
		jo.Wait()

		// fmt.Println("Gorrutina: ", k, " Esperando eBordeIzquierdo")
		eBordeIzquierdo := <-entrada
		_ = eBordeIzquierdo
		// fmt.Println("Gorrutina: ", k, " Recibio eBordeIzquierdo")

		// fmt.Println("Gorrutina: ", k, " Mandando sBordeIzquierdo")
		salida <- sBordeIzquierdo
		// fmt.Println("Gorrutina: ", k, " Envio sBordeIzquierdo")

		for i := range newMapa {
			for j := range newMapa[i] {
				eBordeIzquierdo[i] = append(eBordeIzquierdo[i], newMapa[i][j])
			}
		}
		// for i := 1; i < len(eBordeIzquierdo); i++ {
		// 	for j := 0; j < len(newMapa[i]); j++ {
		// 		newMapa[i][j] = vecinos(eBordeIzquierdo, i, j+1)
		// 	}
		// }

		// renderizar(newMapa)

		mundo.MAPA = newMapa
		// fmt.Println("Gorrutina: ", k, " Mandando mensaje por el channel resultado")
		resultado <- mundo
		// fmt.Println("Gorrutina: ", k, " Mensaje enviado por el channel resultado")
		wg.Done()
	} else {
		var entradaIzquierda chan [][]bool
		var salidaIzquierda chan [][]bool
		var entradaDerecha chan [][]bool
		var salidaDerecha chan [][]bool

		salidaIzquierda = chans[k*2-2]
		entradaIzquierda = chans[k*2-1]
		entradaDerecha = chans[k*2]
		salidaDerecha = chans[k*2+1]

		_, _, _, _ = entradaIzquierda, salidaIzquierda, entradaDerecha, salidaDerecha

		viejoMapa := mundo.MAPA

		newMapa := make([][]bool, len(viejoMapa))
		for i := range viejoMapa {
			newMapa[i] = make([]bool, len(viejoMapa[i]))
			copy(newMapa[i], viejoMapa[i])
		}

		sBordeDerecho := make([][]bool, len(newMapa))
		for i := range sBordeDerecho {
			sBordeDerecho[i] = make([]bool, 1)
		}

		for i := 0; i < len(newMapa); i++ {
			for j := 0; j < 1; j++ {
				sBordeDerecho[i][j] = newMapa[i][len(newMapa[i])-1]
			}
		}

		sBordeIzquierdo := make([][]bool, len(newMapa))
		for i := range newMapa {
			sBordeIzquierdo[i] = make([]bool, 1)
			copy(sBordeIzquierdo[i], mundo.MAPA[i])
		}

		jo.Done()
		// fmt.Println("Gorrutina: ", k, " Before jo.Wait()")
		jo.Wait()

		// fmt.Println("Gorrutina: ", k, " Mandando sBordeDerecho Intermedio")
		salidaDerecha <- sBordeDerecho

		// fmt.Println("Gorrutina: ", k, " Esperando eBordeDerecho")
		eBordeDerecho := <-entradaDerecha
		// fmt.Println("Gorrutina: ", k, " Recibio eBordeDerecho")

		_ = eBordeDerecho

		// fmt.Println("Gorrutina: ", k, " Mandando sBordeIzquierdo Intermedio")
		salidaIzquierda <- sBordeIzquierdo

		// fmt.Println("Gorrutina: ", k, " Esperando eBordeIzquierdo")
		eBordeIzquierdo := <-entradaIzquierda
		// fmt.Println("Gorrutina: ", k, " Recibio eBordeIzquierdo")

		_ = eBordeIzquierdo

		mapaExtendido := make([][]bool, len(newMapa))
		for i := range newMapa {
			mapaExtendido[i] = make([]bool, len(newMapa[i]))
			copy(mapaExtendido[i], newMapa[i])
		}

		// mapaExtendido = append(eBordeDerecho, mapaExtendido...)
		// mapaExtendido = append(mapaExtendido, eBordeIzquierdo...)

		// renderizar(mapaExtendido)

		mundo.MAPA = newMapa
		// fmt.Println("Gorrutina: ", k, " Mandando mensaje por el channel resultado")
		resultado <- mundo
		// fmt.Println("Gorrutina: ", k, " Mensaje enviado por el channel resultado")
		wg.Done()
	}

}

func vecinos(mapa [][]bool, i, j int) bool {
	filas := len(mapa) - 1
	columnas := len(mapa[0]) - 1

	// fmt.Println("Filas: ", filas, "\n", "Columnas: ", columnas)
	con := 0

	if i != 0 && mapa[i-1][j] { // 				↓
		con++
	}
	if i != 0 && j != 0 && mapa[i-1][j-1] { // 	↙
		con++
	}
	if j != 0 && mapa[i][j-1] { // 				←
		con++
	}
	if j != 0 && i != filas && mapa[i+1][j-1] { // 	↖
		con++
	}
	if i != filas && mapa[i+1][j] { // 				↑
		con++
	}
	if j != columnas && i != filas && mapa[i+1][j+1] { // 	↗
		con++
	}
	if j != columnas && mapa[i][j+1] { //				→
		con++
	}
	if j != columnas && i != 0 && mapa[i-1][j+1] { //	↘
		con++
	}
	fmt.Print("Filas: ", filas, " Columnas: ", columnas, " | [", i, "]", "[", j, "]", " CON = ", con, " |")
	return transiciones(mapa[i][j], con)
}

// CALCULO DE LOS NUEVOS ESTADOS: (SE REALIZA DENTRO DEL PROPIO IF)
// Extension fantasma de su mapa
// Recorrer el mapa, contar cuantos vecinos vivos tiene cada celda
// Mandar esa informacion a la funcion que calcula si la celda vive o muere
// Actualizar la informacion en un nuevo mapa temporal para no arruinar el calculo de las demas celdas
// Retornar el mapa temporal con la informacion de las celdas actualizadas
// }
