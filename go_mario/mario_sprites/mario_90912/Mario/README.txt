This is a sample to show how custom characters work ingame, as well as what assets you'll need for them. Make sure any custom characters you make has all of the files present here, all under the same filenames and the same formats (.png for image files, .txt for text files, and .ogg for sound files.)



WHERE TO PUT THIS FOLDER:

Look for "AppData/Local/Frickbears3". Go inside, and look for a folder titled "addons". If there's not already a folder titled "addons" in the Frickbears3 folder, create one and put this character's folder inside it. The title of the custom character's folder is what's used for their name ingame, so make sure you title it appropriately!



----- TEXT ASSETS: Be sure you follow the format of the examples, or else your game may crash when loading this dialogue. -----

* opening_dialogue.txt: Stores an array of strings that make up this character's opening monologue when starting the game. The DIALOGUE value stores each individual line of dialogue as a single string in a comma-separated array, while the FIRST_PERSON value is a Boolean (either "true" or "false") that tells the game if the character is speaking this dialogue themselves (true) or if the narrator is speaking for them (false).

* extras_info.txt: Stores the full name, description and devnotes for the character's extras menu page. Each should be a single string. You may want to include the character's debut date to stay consistent with the extras entries in the vanilla game, assuming they come from a pre-existing piece of media.

* README.txt: What you're reading right now. This file is just for educational purposes; It can be left out for your own custom characters!


----- SPRITE ASSETS: All of your sprites should be the exact same resolution as the examples provided; Otherwise they may look stretched ingame. -----

* icon.png: One sprite. Represents your character's face. Most icons ingame are front-facing and symmetrical, but yours doesn't necessarily need to be.

* minigame: Four sprites. Be sure to use the same four colors present in the sample, otherwise the sprites won't be recolored correctly for the different salvage locations ingame.

* outside: One sprite. It's recommended that the white lines and background are kept consistent when making your own version of the sprite.

* portrait: One sprite. Used for the extras menu. Unlike the other sprites, this one can be any resolution you want!

* reflection: Two sprites. The first sprite should represent your character under normal circumstances, while the second sprite should represent your character when they're in danger.

* selection: Two sprites. The second sprite is overlaid the first to create a shading effect, so make sure their silhouettes are the same.

* silhouette: One sprite. Used for the opening cutscene.


----- SOUND ASSETS: Make sure these are all exported as .OGGs. -----

talksound.ogg: Plays repeatedly when your character is speaking.


----- (SPOILERS) ENDSCREEN ASSETS: It's recommended you hand these off to someone else if you want to avoid spoilers, or only make them once you've seen all of the game's endings already. NOTE: Since these are layered on top of the endscreens ingame, you may want to get your hands on how the isolated backgrounds for the endscreens before making these sprites. You can find these in the game's files, OR, if you want to avoid spoilers altogether, create a completely blank guard, get the endings as them, and screenshot the endscreens once you reach them to use as a base. -----

* evil.png: One sprite. Used for the evil route endscreen. It's recommended you only draw over the parts of the sprite present in the example.

* good.png: One sprite. Used for the good route endscreen. Your character should be positioned near the bottom left of the sprite.

* money.png: One sprite. Used for the money route endscreen. Your character should be positioned to the right, and have a gloved hand on their shoulder.

* slacker.png: One sprite. Used for the slacker route endscreen. Your character should be sitting down.


----- (SPOILERS) CUTSCENE ASSETS: Like the previous set of assets, make sure you've seen every ending before you make these.

* ending_dialogue.txt: Same format as opening_dialogue.txt, but with the addition of a "PHOTO_CAPTION" value. This should be a single string, and acts as the caption for endphoto.png when it appears ingame.

* conversation.png: Two sprites. Used for when your character is talking to someone about half their height.

* endphoto.png: One sprite. Shows what your character gets up to after the events of the story.

* files.png: One sprite. Be sure the space to the left of their head and above the folder remains empty.

* gameover.png: One sprite. Your character after meeting a terrible fate.

* springtrap.png: One sprite. Should be a straight-on, low-angle, symmetrical shot of your character. Make sure the space above them remains empty.
