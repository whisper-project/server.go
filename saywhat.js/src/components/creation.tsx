import React from 'react'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'

import { SettingsView } from './settings'
import { Generate } from './generate'
import { Settings } from '../model/settings'

export function Creation(props: { settings: Settings }) {
    return (
        <Stack spacing={2}>
            <Typography variant="h4" gutterBottom>
                Generation
            </Typography>
            <Generate settings={props.settings} />
            <Typography variant="h4" gutterBottom>
                Settings
            </Typography>
            <SettingsView settings={props.settings} />
        </Stack>
    )
}
