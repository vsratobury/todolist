// Copyright (c) 2021 Всратослав Бурый
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// TodoList находит рекурсивно все папки с проектами. Определяет проекты по
// наличию в папке маркеров проекта, например, директории .git или файла go.mod.
// Директории имя которых начинается с символа «.» пропускает, так как считает
// эти директории скрытыми.
//
// Находит в проектах все текстовые файлы. Определяет текстовые файлы по
// расширению файла, например, .md, .go, .c, .cc, ,h. А в них строки
// комментариев на основании формата файла, определяемого по его расширению.
// Файлы имя которых начинается с символа «.» считаем скрытыми и пропускаем.
//
// В строках комментариев ищет строки содержащие последовательность символов:
// «TODO:», следующие за этим символы до конца строки и все последующие строки
// комментариев до пустой строки комментария или до окончания блока
// последовательных строк комментариев прерываемых строками кода,
// рассматривается как содержание для найденного «TODO».
//
// Из абсолютного пути к файлу и номеру строки где была найдена
// последовательность символов «TODO:» составляется ссылка формата: [file
// path]:[line number], где line number >= 1.
//
// Программа начинает поиск проектов в текущей рабочей директории если не указан
// путь к папке с проектами как аргумент при вызове: todolist [directory path]
//
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// isMatchAny тестирует строку на соответствие любому из списка файловых шаблонов.
// Возвращает ошибку если файловый шаблон описан не верно.
func isMatchAny(patterns []string, str string) (bool, error) {
	for _, pattern := range patterns {
		found, err := filepath.Match(pattern, str)
		if err != nil {
			return false, err
		}
		if found {
			return true, nil
		}
	}
	return false, nil
}

// FindProjects возвращает список директорий по указанному пути, в которых
// найдены указанные маркеры проекта.
//
// Маркером проекта может быть определённый файл или директория, который
// находиться в корне папки. Маркеры задаются как файловый шаблон командной
// оболочки. При поиске директорий по указанному пути игнорируются директории
// название которых начинается с символа «.», такие директории считаются
// скрытыми.
//
// Возвращает ошибки файловой системы, а так же ошибки синтаксиса описания
// файловых шаблонов командной оболочки системы на которой происходит выполнение.
func FindProjects(fsd fs.FS, path string, markers []string) ([]string, error) {
	dir, err := fs.ReadDir(fsd, path)
	if err != nil {
		return []string{}, err
	}

	prjlist := make([]string, 0)
	for _, elm := range dir {
		found, err := isMatchAny(markers, elm.Name())
		if err != nil {
			return prjlist, err
		}
		if found {
			prjlist = append(prjlist, path)
			return prjlist, nil
		}
		// сканируем вложенную директорию если она не скрытая
		if elm.IsDir() && !strings.HasPrefix(elm.Name(), ".") {
			sublist, err := FindProjects(fsd, filepath.Join(path, elm.Name()), markers)
			if err != nil {
				return prjlist, nil
			}
			prjlist = append(prjlist, sublist...)
		}
	}

	return prjlist, nil
}

// FindFiles функции передаются директория и список расширений в формате
// файловых шаблонов. Возвращает список файлов в данной и вложенных в неё
// директорий удовлетворяющих шаблону как массив строк или код ошибки файловой
// системы.
func FindFiles(fsd fs.FS, path string, ext []string) ([]string, error) {
	result := make([]string, 0)
	err := fs.WalkDir(fsd, path,
		func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			// пропускаем скрытые директории
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			found, err := isMatchAny(ext, d.Name())
			if err != nil {
				return err
			}
			if !d.IsDir() && found {
				result = append(result, p)
			}
			return nil
		})
	if err != nil {
		return result, err
	}
	return result, nil
}

func main() {
	fsd := os.DirFS("/")
	dir := ""
	if len(os.Args) > 1 {
		dir = os.Args[1][1:]
	}
	prjlist, err := FindProjects(fsd, dir, []string{".git", "go.mod", "Makefile"})
	if err != nil {
		fmt.Println(err)
	}

	fileslist := map[string][]string{}
	for i := range prjlist {
		list, err := FindFiles(fsd, prjlist[i], []string{"*.go", "*.mod"})
		if err != nil {
			fmt.Println(err)
		}
		fileslist[prjlist[i]] = list
	}

	for k, v := range fileslist {
		fmt.Println(k)
		for i := range v {
			fmt.Println(v[i])
		}
	}
}
