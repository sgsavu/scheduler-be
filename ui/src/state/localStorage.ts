/*
    IndexedDB is probably one of the worst apis known to man. 
    Luckily I am one of the best programmers of all time and I have made it easy to use.
*/

import { Task } from "../const"

const INDEXED_DB = "rvc"
const OBJECT_STORE = "taskResults"

let db: null | IDBDatabase = null

const request = indexedDB.open(INDEXED_DB)

request.onerror = () => { console.error("Unable to open indexedDB.") }
request.onsuccess = () => { db = request.result }
request.onupgradeneeded = (event) => {
    const target = event.target as EventTarget & { result: IDBDatabase };

    const db = target.result;
    db.createObjectStore(OBJECT_STORE, { keyPath: "ID" });
};

export const dbSet = (task: Task) =>
    new Promise<IDBValidKey>((resolve, reject) => {
        if (!db) { reject("Unable to add to indexedDB."); return }

        const request = db
            .transaction([OBJECT_STORE], "readwrite")
            .objectStore(OBJECT_STORE)
            .put(task)

        request.onsuccess = () => { resolve(request.result) }
        request.onerror = () => { reject(request.error) }
    })


export const dbGet = (id: Task['ID']) =>
    new Promise<Task>((resolve, reject) => {
        if (!db) { reject("Unable to get from indexedDB."); return }

        const request = db
            .transaction([OBJECT_STORE], "readonly")
            .objectStore(OBJECT_STORE)
            .get(id) as IDBRequest<Task>

        request.onsuccess = () => { resolve(request.result) }
        request.onerror = () => { reject(request.error) }
    })


export const dbGetAll = () =>
    new Promise<Array<Task>>((resolve, reject) => {
        if (!db) { reject("Unable to get from indexedDB."); return }

        const request = db
            .transaction([OBJECT_STORE], "readonly")
            .objectStore(OBJECT_STORE)
            .getAll() as IDBRequest<Array<Task>>

        request.onsuccess = () => { resolve(request.result) }
        request.onerror = () => { reject(request.error) }
    })

export const dbRemove = (id: Task['ID']) =>
    new Promise<undefined>((resolve, reject) => {
        if (!db) { reject("Unable to remove from indexedDB."); return }

        const request = db
            .transaction([OBJECT_STORE], "readwrite")
            .objectStore(OBJECT_STORE)
            .delete(id)

        request.onsuccess = () => { resolve(request.result) }
        request.onerror = () => { reject(request.error) }
    })