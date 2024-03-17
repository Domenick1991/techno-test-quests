package auth

import (
	"net/http"
	users "techno-test_quests/quests/handlers/user"
	"techno-test_quests/quests/storage"
)

// NonPage handler для пустой страницы
func NonPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Страница не существует"))
}

// AdminAuth Авторизация администратора
func AdminAuth(next http.HandlerFunc, storage *storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			user, err := users.GetUser(username, password, storage)
			if err != nil {
				http.Error(w, "Ошибка при проверки пользователя", http.StatusBadRequest)
				return
			}
			if user.Isadmin {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Введен неверный логин/пароль", http.StatusUnauthorized)
	})
}

// UserAuth Авторизация любого пользователя
func UserAuth(next http.HandlerFunc, storage *storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			_, err := users.GetUser(username, password, storage)
			if err != nil {
				http.Error(w, "Ошибка при проверки пользователя", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Введен неверный логин/пароль", http.StatusUnauthorized)
	})
}
