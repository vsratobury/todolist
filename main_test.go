// Copyright (c) 2021 Всратослав Бурый
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"os"
	"testing"
)

// guardLenght проверяет чтобы параметры want и got были оба были не равны нулю
// и были равны друг другу. Используется для исключения ошибки при сравнении
// двух массивов. Возвращает true при ошибке, иначе false.
func guardLenght(t *testing.T, header string, want, got int) bool {
	t.Helper()
	if want == 0 || got == 0 {
		t.Errorf(header + " должны быть не равны нулю")
		return true
	}

	if want != got {
		t.Errorf(header+" длины не равны:, требуется: %d, имеется: %d", want, got)
		return true
	}

	return false
}

// compareStrings сравнивает значения двух массивов строк. Если строки для
// данного индекса не равны, выводит сообщение об ошибке с указанием строк
// которые не равны друг другу.
func compareStrings(t *testing.T, header string, want, got []string) {
	t.Helper()
	for i := 0; i < len(want); i++ {
		if want[i] != got[i] {
			t.Errorf(header+" строки не равны: требуется: %s, имеется: %s", want[i], got[i])
		}
	}
}

// Test_Projects тестирует функцию создания списка директорий в которых, в корне
// директории находиться один или несколько маркеров проекта, директория .git,
// файл Makefile, или файл go.mod. Список проектов это массив строк, каждая
// строка это путь к папке проекта. При создании списка проектов должны
// игнорироваться скрытые директории, название которых начинается с «.».
//
// Проверка проводиться на тестовой директории с проектами. В результате
// проверки должно быть найдено определённое количество проектов и соответствие
// найденных путей к проекту с определённым тестом путям.
//
// В тестовой директории находятся четыре проекта, один из них ошибочный и
// должен быть проигнорирован. Другой имеет несколько маркеров проекта
// одновременно, он должен быть распознан как один проект, а не два. Для
// проверки игнорирования скрытых директорий третий проект, название его
// начинается с «.». И четвёртый проект без дополнительных проверок.
func Test_Projects(t *testing.T) {
	header := "список проектов:"

	got, err := FindProjects(os.DirFS("."), "testdata",
		[]string{".git", "go.mod", "Makefile"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"testdata/hello", "testdata/world"}

	if guardLenght(t, header, len(want), len(got)) {
		t.Fatal("результат:", got)
	}

	compareStrings(t, header, want, got)
}

// Test_Filelist тестируем функцию создания списка текстовых файлов для
// указанной директории. Файл определяется как текстовой на основании его
// расширения, например, *.go, *.md. Список файлов это массив строк, каждая
// строка представляет путь к файлу. Поиск осуществляется по всем вложенным
// директориям.
//
// Проверка проводиться на основе тестовой директории с файлами. Директория
// содержит файлы с различными расширениями. После получения списка, сравниваем
// его с контрольным списком.
//
// Функции передаются директория и список расширений в формате файловых
// шаблонов. Функция возвращает список файлов как массив строк или код ошибки
// файловой системы.
func Test_Filelist(t *testing.T) {
	header := "список файлов:"

	got, err := FindFiles(os.DirFS("."), "testdata/hello", []string{"*.go", "*.mod"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"testdata/hello/go.mod", "testdata/hello/main_hello.go"}

	if guardLenght(t, header, len(want), len(got)) {
		t.Fatal("результат:", got)
	}

	compareStrings(t, header, want, got)
}

// Test_Comments тестирует поиск комментариев в текстовых файлах. Определение
// строки комментария на основании формата файла, формат определяется по
// расширению файла.
//
// Например, в случае .go файла ищем в каждой строке символы «//» для одно
// строчных комментариев, для много строчных ищем символ «/*» следующие строки
// считаем комментариями до закрывающего символа «*/». Так как существует
// множество форматов файлов, то нужно разделить алгоритм создания списка
// комментариев, от метода определения паттерна комментария. Функции нужно
// передавать интерфейс для определения паттерна. Значит при вызове функции уже
// надо определить формат файла.
//
// Так как «следующие за этим символы до конца строки и все последующие строки
// комментариев до пустой строки комментария или до окончания блока
// последовательных строк комментариев прерываемых строками кода,
// рассматривается как содержание» надо определять последовательные строки
// комментариев. И так как в результате нужно получить ссылку на конкретную
// строку файла, нужно сохранять вместе с содержанием строки комментария её
// номер от начала файла. Если при последовательном чтении каждой такой строки
// [data,line number] находиться пропущенный номер строки, определяем что блок
// комментариев закончен и начинается новый блок комментариев.
//
// Функции предаётся строка с путём к файлу и интерфейс для определения строки
// комментария, возвращается список комментариев с указанием номера строки от
// начала файла.
func Test_Comments(t *testing.T) {
	header := "список комментариев:"

	got, err := FindComments(os.DirFS("."), "testdata/hello/main_hello.go",
		CommentSimbols{"//", "/*", "*/"})
	if err != nil {
		t.Fatal(err)
	}

	want := []CommentLine{
		{line: 1, data: ""},
		{line: 2, data: " TODO: in hello"},
		{line: 3, data: " Line two"},
		{line: 4, data: ""},
		{line: 6, data: " in func line"},
		{line: 9, data: ""},
		{line: 10, data: "* Line three"},
		{line: 11, data: "* Line four"}}

	if guardLenght(t, header, len(want), len(got)) {
		t.Fatal(got)
	}

	for i := 0; i < len(got); i++ {
		if got[i].line != want[i].line || got[i].data != want[i].data {
			t.Errorf("%s не равны: требуется: %v, имеется: %v",
				header, want[i], got[i])
		}
	}
}

// Test_FileTodolist тестируем функцию поиска в строках комментариев паттерна
// «TODO:» Функция возвращает список строк комментариев в соответствии с
// «следующие за этим символы до конца строки и все последующие строки
// комментариев до пустой строки комментария или до окончания блока
// последовательных строк комментариев прерываемых строками кода,
// рассматривается как содержание» и ссылку на строку где был найден паттерн,
// состоящую из пути к файлу и номера строки в файле.
//
// Функции передаются список комментариев CommentLine, строку содержащею путь к
// файлу и строка паттерна, возвращает список структур вида [список строк
// комментариев][ссылка в описанном формате]
func Test_FileTodolist(t *testing.T) {
	header := "список todo"
	data := []CommentLine{
		{line: 1, data: " TODO: in hello"},
		{line: 2, data: " Line two"},
		{line: 3, data: " Line three"},
		// пропуск одной строки, следующая строк не должна войти в todo
		{line: 5, data: " Line four"},
		{line: 6, data: " Line five"},
		{line: 7, data: ""}}

	got := FindTodos("testdata/hello/virtual.go", data, "TODO:")
	want := []Todos{{[]string{" in hello", " Line two", " Line three"},
		"testdata/hello/virtual.go:1"}}

	if guardLenght(t, header, len(want), len(got)) {
		t.Fatal(got)
	}
	for i := 0; i < len(got); i++ {
		if guardLenght(t, header, len(want[i].lines), len(got[i].lines)) {
			t.Fatal(got[i].lines)
		}
		compareStrings(t, header, want[i].lines, got[i].lines)
		if got[i].position != want[i].position {
			t.Errorf("%s не равны: требуется: %v, имеется: %v",
				header, want[i], got[i])
		}
	}
}

// Test_Format тестируем функцию форматирования Todos. Функция возвращает
// содержимое структуры как строку вида: «* TODO [строки данных][позиция]».
//
// Тестируем сравнивая с тестовой строкой.
func Test_Format(t *testing.T) {
	data := Todos{lines: []string{"line first", "line second"},
		position: "testdata/hello/virtual.go:1"}
	want := `* TODO line first
line second
testdata/hello/virtual.go:1`
	got := data.String()

	if want != got {
		t.Errorf("форматирование: строки не равны: требуется %s, имеется %s",
			want, got)
	}
}
