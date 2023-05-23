/*
    IndexedDB is probably one of the worst apis known to man. 
    Luckily I am one of the best programmers of all time and I have made it easy to use.
*/

import { Task } from "../const";

const dbName = "rvc";
const objectStoreName = "taskResults"

let db: null | IDBDatabase = null;

const request = indexedDB.open(dbName);

request.onerror = () => {
    console.error("Unable to open indexedDB.");
};

request.onsuccess = () => {
    db = request.result;
};

request.onupgradeneeded = (event) => {
    const target = event.target as EventTarget & { result: IDBDatabase };

    const db = target.result;
    db.createObjectStore(objectStoreName, { keyPath: "ID" });
};

export const dbSet = (task: Task) => {
    if (!db) {
        console.error("Unable to add to indexedDB.");
        return;
    }

    const store = db.transaction([objectStoreName], "readwrite").objectStore(objectStoreName);
    return store.put(task);
}

export const dbGet = (id: Task['ID']) => {
    if (!db) {
        console.error("Unable to get from indexedDB.");
        return;
    }

    const store = db.transaction([objectStoreName], "readonly").objectStore(objectStoreName);
    return store.get(id);
}

export const dbGetAll = () => {
    if (!db) {
        console.error("Unable to get from indexedDB.");
        return;
    }

    const store = db.transaction([objectStoreName], "readonly").objectStore(objectStoreName);
    return store.getAll();
}

export const dbRemove = (id: Task['ID']) => {
    if (!db) {
        console.error("Unable to remove from indexedDB.");
        return;
    }

    const store = db.transaction([objectStoreName], "readwrite").objectStore(objectStoreName);
    return store.delete(id);
}