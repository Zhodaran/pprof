package pprof

import (
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"net/http"
	_ "net/http/pprof"
)

func CreatePprof() {

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	cpuProfileFile, err := os.Create("profile.prof")
	if err != nil {
		log.Fatal("Не удалось создать файл для CPU-профиля:", err)
	}
	defer cpuProfileFile.Close()

	// Начинаем профилирование CPU
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		log.Fatal("Не удалось начать CPU-профилирование:", err)
	}
	defer pprof.StopCPUProfile()

	traceFile, err := os.Create("trace.out")
	if err != nil {
		log.Fatal("Не удалось создать файл для трассировки:", err)
	}
	defer traceFile.Close()

	// Начинаем трассировку
	if err := trace.Start(traceFile); err != nil {
		log.Fatal("Не удалось начать трассировку:", err)
	}
	defer trace.Stop()
}
