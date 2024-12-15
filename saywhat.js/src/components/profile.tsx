/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

import React, { useEffect } from 'react'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'

import { Profile, SettingsStore } from '../model/settings'

const whisperUrl = 'https://apps.apple.com/app/whisper-talk-without-voice/id6446479064'

export function ProfileView(props: { profile: Profile }) {
    const [profileId, setProfileId] = React.useState(props.profile.id)
    const [password, setPassword] = React.useState(props.profile.password)
    useEffect(() => {
        SettingsStore.updateLocalProfile(profileId, password)
    }, [profileId, password])
    function onProfileChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
        setProfileId(e.currentTarget.value)
    }
    function onPasswordChange(e: React.ChangeEvent<HTMLInputElement>) {
        setPassword(e.currentTarget.value)
    }
    function updateProfile() {
        SettingsStore.updateLocalProfile(profileId, password)
    }
    function createNewProfile() {
        const id = crypto.randomUUID().toUpperCase()
        const password = generatePassword()
        SettingsStore.updateLocalProfile(id, password)
    }
    return (
        <Stack spacing={2}>
            <Typography variant={'h2'} alignItems="center" gutterBottom>
                Profile Sharing
            </Typography>
            <Typography>
                This application lets you experiment with ElevenLabs speech settings for and edit
                your Favorites in <a href={whisperUrl}>the Whisper app</a>.
            </Typography>
            <Typography>
                To use an existing shared profile, enter the profile ID and password in the boxes
                below, and hit the "Download Profile" button. To create a new shared profile, hit
                the "Create New Profile" button.
            </Typography>
            <TextField
                id="outlined-basic"
                label="Profile ID"
                variant="outlined"
                value={profileId}
                onChange={onProfileChange}
            />
            <TextField
                id="outlined-basic"
                label="Profile Password"
                variant="outlined"
                value={password}
                onChange={onPasswordChange}
            />
            <Button variant="contained" onClick={updateProfile} disabled={!profileId || !password}>
                Download Profile
            </Button>
            <Button variant="contained" onClick={createNewProfile}>
                Create New Profile
            </Button>
        </Stack>
    )
}

function generatePassword() {
    const letters = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%_+-='
    const anyLetter = () => letters[Math.floor(Math.random() * letters.length)]
    return Array(20).fill(0).map(anyLetter).join('')
}
