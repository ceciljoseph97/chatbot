class AdvancedCommands:
    def __init__(self):
        self.commands = [
            {
                'name': 'Enable Animated Printing',
                'flag': '-anim',
                'type': 'boolean',
                'description': 'Enable or disable animated letter-by-letter printing.'
            },
            {
                'name': 'Set Corpus File',
                'flag': '-c',
                'type': 'string',
                'description': 'The file to store corpora (default "PMFuncOverView.gob").'
            },
            {
                'name': 'Set CMEM',
                'flag': '-cmem',
                'type': 'integer',
                'description': 'Number of conversations the context remains active (2-4) (default 2).'
            },
            {
                'name': 'Set Config File',
                'flag': '-config',
                'type': 'string',
                'description': 'Path to the config file (default "./config_local.yaml").'
            },
            {
                'name': 'Enable Context Handling',
                'flag': '-context',
                'type': 'boolean',
                'description': 'Enable or disable context handling (default true).'
            },
            {
                'name': 'Enable Developer Mode',
                'flag': '-dev',
                'type': 'boolean',
                'description': 'Enable developer mode.'
            },
            {
                'name': 'Show Intro Message',
                'flag': '-intro',
                'type': 'boolean',
                'description': 'Show the intro message (default true).'
            },
            {
                'name': 'Set Number of Answers',
                'flag': '-t',
                'type': 'integer',
                'description': 'The number of answers to return.'
            },
        ]

    def get_commands(self):
        return self.commands
