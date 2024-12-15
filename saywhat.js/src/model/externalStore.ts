/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import { SettingsStore } from './settings'

export class ApiExternalStore<T> {
    callbacks: (() => void)[] = []
    cacheData: T[] = []
    fetchData: () => Promise<T[]>

    constructor(fetchData: () => Promise<T[]>) {
        this.fetchData = fetchData
    }

    subscribe(callback: () => void) {
        console.log(`Subscribed to ${this.constructor.name} Store`)
        this.callbacks = [...this.callbacks, callback]
        let unsubscribeKeyChange = SettingsStore.subscribeKeyChange(() => this.doFetch)
        if (
            this.cacheData.length == 0 &&
            SettingsStore.getSnapshot().settings.api_key.length >= 32
        ) {
            this.doFetch()
        }
        return () => {
            console.log(`Unsubscribed from ${this.constructor.name} Store`)
            this.callbacks = this.callbacks.filter((c) => c !== callback)
            unsubscribeKeyChange()
        }
    }

    notify() {
        this.callbacks.map((c) => c())
    }

    getSnapshot() {
        return this.cacheData
    }

    doFetch() {
        if (SettingsStore.getSnapshot().settings.api_key.length < 32) {
            console.warn(`Invalid API key, clearing ${this.constructor.name} Store cache`)
            this.cacheData = []
            this.notify()
        } else {
            console.log(`Performing API fetch for type ${this.constructor.name}...`)
            this.fetchData().then((data) => {
                console.log(
                    `Fetched ${this.cacheData.length} objects of type ${this.constructor.name}`,
                )
                this.cacheData = data
                this.notify()
            })
        }
    }
}
