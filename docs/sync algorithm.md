# Алгоритм синхронизации данных

## Матрица решений для записей, существующих на обеих сторонах
> sd - server data
>
> cd - client data

| `sd.Version` vs `cd.Version` | `sd.IsDeleted` | `cd.IsDeleted` | `sd.Hash == cd.Hash` | Действие |
|---|---|---|---|---|
| `==` | `true` | `true` | — | ничего |
| `==` | `true` | `false` | — | удалить на клиенте |
| `==` | `false` | `true` | — | удалить на сервере |
| `==` | `false` | `false` | `true` | ничего |
| `==` | `false` | `false` | `false` | обновить на сервере |
| `sf > cf` | `true` | — | — | удалить на клиенте |
| `sf > cf` | `false` | — | — | скачать с сервера |
| `sf < cf` | — | `true` | — | удалить на сервере |
| `sf < cf` | — | `false` | — | обновить на сервере |

## Матрица для записей только на одной стороне

| Только на сервере `sd.IsDeleted` | Только на клиенте `cd.IsDeleted` | Действие |
|---|---|---|
| `false` | — | скачать с сервера |
| `true` | — | ничего (удалена до первой синхронизации) |
| — | `false` | загрузить на сервер |
| — | `true` | ничего (создана и удалена локально) |

---

## Алгоритм

```go
func SyncDecision(
    serverData []PrivateDataState,
    clientData []PrivateDataState,
) SyncPlan {

    var plan SyncPlan // Download, Upload, Update, DeleteClient, DeleteServer

    clientIndex := make(map[string]PrivateDataState, len(clientData))
    for _, cd := range clientData {
        clientIndex[cd.ClientSideID] = cd
    }

    serverIndex := make(map[string]PrivateDataState, len(serverData))
    for _, sd := range serverData {
        serverIndex[sd.ClientSideID] = sd
    }

    // ── Шаг 1: проход по серверным записям ──────────────────────────────────
    for _, sd := range serverData {
        cd, existsOnClient := clientIndex[sd.ClientSideID]

        if !existsOnClient {
            if !sd.IsDeleted {
                // Новая запись на сервере → скачать на клиент
                plan.Download = append(plan.Download, sd)
            }
            // sd.IsDeleted && !existsOnClient → ничего: удалена до первой синхронизации
            continue
        }

        // Запись есть на обеих сторонах
        switch {
        case sd.Version == cd.Version:
            switch {
            case sd.IsDeleted && cd.IsDeleted:
                // Обе стороны удалили → синхронизировано

            case sd.IsDeleted && !cd.IsDeleted:
                // Сервер удалил, клиент не знает → удалить на клиенте
                plan.DeleteClient = append(plan.DeleteClient, sd)

            case !sd.IsDeleted && cd.IsDeleted:
                // Клиент удалил, сервер не знает → удалить на сервере
                plan.DeleteServer = append(plan.DeleteServer, cd)

            case sd.Hash == cd.Hash:
                // Версия и хэш совпали → данные идентичны, ничего не делаем

            default: // !sd.IsDeleted && !cd.IsDeleted && sd.Hash != cd.Hash
                // Одинаковая версия, хэши различаются →
                // клиент изменил данные локально → обновить на сервере
                plan.Update = append(plan.Update, cd)
            }

        case sd.Version > cd.Version:
            // Сервер опережает клиента
            if sd.IsDeleted {
                // Сервер удалил более новую версию → удалить на клиенте
                plan.DeleteClient = append(plan.DeleteClient, sd)
            } else {
                // На сервере более актуальная версия → скачать на клиент
                plan.Download = append(plan.Download, sd)
            }

        default: // sd.Version < cd.Version
            // Клиент опережает сервер (offline-изменения)
            if cd.IsDeleted {
                // Клиент удалил более новую версию → удалить на сервере
                plan.DeleteServer = append(plan.DeleteServer, cd)
            } else {
                // Клиент имеет более актуальные изменения → обновить на сервере
                plan.Update = append(plan.Update, cd)
            }
        }
    }

    // ── Шаг 2: проход по клиентским записям, которых нет на сервере ─────────
    for _, cd := range clientData {
        if _, existsOnServer := serverIndex[cd.ClientSideID]; existsOnServer {
            continue // уже обработано в шаге 1
        }

        if !cd.IsDeleted {
            // Новая запись только на клиенте → загрузить на сервер
            plan.Upload = append(plan.Upload, cd)
        }
        // cd.IsDeleted && !existsOnServer → ничего:
        // создана и удалена локально, на сервер никогда не попадала
    }

    return plan
}

```
