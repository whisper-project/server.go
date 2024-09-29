import React, { useSyncExternalStore } from 'react'
import Grid from '@mui/material/Grid'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'

import { Creation } from './components/creation'
import { SettingsStore } from './model/settings'
import { History } from './components/history'
import { ProfileView } from './components/profile'

export function App() {
    const { profile, settings } = useSyncExternalStore(
        SettingsStore.subscribe,
        SettingsStore.getSnapshot,
    )
    return (
        <Grid container spacing={4}>
            <Grid item xs={4}>
                <ProfileView key={SettingsStore.profile.password} profile={profile} />
            </Grid>
            <Grid item xs={4}>
                <Creation key={SettingsStore.eTag} settings={settings} />
            </Grid>
            <Grid item xs={4}>
                <History key={SettingsStore.eTag} settings={settings} />
            </Grid>
        </Grid>
    )
}
