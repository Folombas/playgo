Well, look who it is! This is the spoiler version of README.txt.
Everything tackled here will contain spoilers for the full game! If you haven't seen everything and care for spoilers, run for the hills for there is no secret spared here!
Most preliminary information is explained in the regular version of this file, so be sure to read over that one first to understand everything here!

== TEXT ASSETS ==

	- SPOILER_README.txt
		The joke was funny the first time. I'm not saying it again.
	
	- *ending_dialogue.txt
		The dialogue for the guard when speaking to Cassidy during the Ultimate route. 
		This file will not be referenced if the "new_cassidy_dialogue" property is set to true, but you should still include this file for compatibility with older versions as well as the credits!
		The "DIALOGUE" value should store each line of dialogue in a comma-seperated array, and the "FIRST_PERSON" value should be a true/false boolean which determines if it's the guard or the narrator speaking.
		Additionally, "PHOTO_CAPTION" should be a single string that contains the flavor text for the guard's photograph in the ending's credits.


== IMAGE ASSETS ==

	- *springtrap.png
		A 480x232 image of the guard at a straight-on, low angle, symmetrical shot.
		The space above them should remain free, and if their outline is white, should have some of it cut off for better blending.
	
	- heldfiles.png
		A 585x232 image that is a three-frame loop of the guard holding Talbert's files. Can be removed if your guard's hand is visually similar to a regular hand.
	
	- *files.png
		A 480x232 image of the guard at a low angle looking down at the afformentioned files.
	
	- *gameover.png
		A 240x232 image that is a grayscale drawing of the guard's corpse, with their bloodied face covered by a cartoon skull.
	
	- *endphoto.png
		A 280x360 image depicting a polaroid photo of your guard after the events of the Ultimate route. Where are they now? You decide!
	
	- slacker.png, evil.png, good.png & money.png
		All 640x360 images depicting the guard in each respective ending.
		
		Slacker should depict them lounging at home, the dim glow of the television illuminating them spending their free time... however they are.
		
		Evil should depict them in the corner as one of the victims of the fire, with their name mentioned in the obituaries. Use Times New Roman for the font, or find some clever way to cover up where the name should be!

		Good should depict them looking up at the sunrise -- and more specifically -- the balloons floating away.
		
		Money should depict them sitting uncomfortably at the dinner table, with the Boss's hand on their shoulder or back.
		
		Of course, feel free to break these conventions however you want! They're merely what's *usually* happening, go crazy with it if you'd like!


== PROPERTIES ==
	All properties here are formatted the same as usual, but must be put in the appropriate "spoiler_properties.txt" file.
	
	The full list of properties this document can set is as follows:
	
	- "theme_victory_night6" | Array | Default: [sfxWarioWinUltimate, 0.5]
		Same as the other victory themes, but for clearing night six. As before, this replaces the track "6:00 AM (Night 6)", not the bells.
	
	- "new_cassidy_dialogue" | Boolean | Default: false
		If the value is true, the game will check in the "dialogue" folder for "cutscene_Cassidy.txt", and will use that file for the dialogue when the guard starts talking to him in the Ultimate ending.
		If false, the game will use "ending_dialogue.txt". You can use this property if you want your guard to have more dialogue options compared to the old system, or if you want Cassidy to speak with the guard.
	
	- "molten_freddy" | Array
		Dialogue for after Molten Freddy reveals themself from the locked box.
	
	- "circus_baby" | Array
		Dialogue for the guard's reaction to Circus Baby nonchalantly and falsely stating that they would salvage her.
	
	- "animdude" | Array
		Dialogue for the awkward discussion between the guard and Animdude.
	
	- "popgoes" | Array
		"If I did, would that make you myPOPGOES?"
	
	- "rodney" | Array
		Dialogue for the guard's disbelief of Rodney's alleged fame.


== DIALOGUE ==

	- ["_prescripted_sequence", <scene>]
		Allows you to play one of two pre-scripted scenes, both for the Ultimate route. They are as follows:
		- "mci"
			Plays the scene where Vanny explains the Missing Children Incident on night four.
		- "talbert"
			Plays the scene where the three supporting characters argue amongst themselves on
			night five (starting with the guard's usual last line of dialogue before this occurs.)
	
	- ["_split", <isGlobalScope?>, <dataName>, <valueToCompare>, <isEqualPath>, <isDifferentPath>]
		Perhaps the second-most complex command. Actually it might be the most complex, I just didn't want to spoil you. Sorry.
		Unless you feel confident in using this command, I'd recommend not messing with it.
		All story branch splits in-game are already included in the template scripts, so you can just edit their text as needed.
		But who would I be if I didn't at least try to explain it?
		
		This command allows you to create branching paths in the script based on story conditions.
		Nested splits do work! So be sure to use those if need-be.
		
		<isGlobalScope?> is a boolean which determines if the data referenced in the next argument is meant to be a global varaible.
			If not, it will pull from a preset struct, explained in the following argument.
		
		<dataName> is the name of the variable or data to compare to the following argument as a string.
			If the previous argument is true, then this must be a reference to any global variable, such as "global.Night".
			If not, this must reference the name of one of the following preset data values, provided as follows:
			- "IsGood": Is the player currently on the Good route? (Collected all masks up to this night.)
			- "IsEvil": Is the player currently on the Evil route? (Salvaged all of Vanny's animatronics up to this night.)
			- "IsUltimate: Is the player currently on the Ultimate route? (On the Good route and the Evil route. Money route is not checked.)
			- "WasGood": Was the player previously on the Good route? (Was collecting all masks, but not necessarily still collecting them.)
			- "WasEvil": Was the player previously on the Evil route? (Was salvaging all of Vanny's animatronics, but not necessarily still salvaging them.)
			- "WasUltimate": Was the player previously on the Ultimate route? (Was on the Ultimate route, but has since stopped collecting both sets of required items.)
			- "HasTalbert": Does the player currently own Talbert's files?
		
		<valueToCompare> is much simpler, it's simply the value to compare (WHAT!) to <dataName>. Can be a boolean or a number.
		
		<isEqualPath> & <isDifferentPath> are the different sets of dialogue that play dependent on how the value of <dataName> and <valueToCompare>, well, compare.
		
		Example:
		["_split", false, "IsEvil", true, 
			[
				["Wow, Ricoh, you've become EVILS!", "Sparks"],
				["_split", true, "Night", 3,
					[
						["On a full moon? It can't be a coincidence... you're a WEREWOLF!!!", "Sparks"],
					],
					
					[
						["Wait, it's not a full moon?... You must not be Ricoh!", "Sparks"],
						["You're... YOU'RE...", "Sparks"],
						["You're HOCIR!!!", "Sparks"],
					]
				],
				["The evils proceeded to leave his body.", ""]
			],
			
			[
				["I hope no evils take over you, Ricoh! That would suck big-time!", "Sparks"]
			]
		]


== DIALOGUE (EXT.) ==
	
	- ["addActor", <pose>, <sprite>]
		Adds an actor to the current cutscene.
		<pose> is a number which determines the pose the actor will begin in.
		<sprite> is a string which is the sprite asset the character should
			draw from (should either be "sTalksprites_Vanny" or "sTalksprites_Michael").
	
	- ["setActor", <pose>]
		Sets the pose for the current actor.
		<pose> is a number which determines the pose the currently focused-on actor will switch to.
	
	- ["switchActor", [<pose>, <focus>], *<sprite>]
		Switches focus to the other actor present in the scene.
		<pose> is a number which determines the pose the actor will switch to.
		<focus> is a number that sets which actor to focus on (should be 0 for the main actor, and 1 for the secondary actor).
		<sprite> is an optional variable which is a string that determines the sprite the actor will have.
			Should only be used when initializing the secondary actor (aka on the first time this function is used).
	
	- ["removeActor"]
		Removes the current actor from the scene.
	
	- ["night6", <route>]
		Goes to night six.
		<route> is a string that refers to the route to enter night six in. Should either be "vanny" or "michael".
	
	- ["talbertReveal"]
		Reveals the player holding Talbert's files as part of the Ultimate night five scene.
	
	- ["talbertPuppet"]
		Ditto, but for when switching to the shot of the Marionette looming over them.
	
	- ["talbert"]
		Goes to the scene where the player reads Talbert's files.

	- ["_money"]
		Goes to the dialogue for when a player meets the qualifications for the money route and another route at the same time.

For all <pose>-related functions, the available poses for the two characters are as follows:

	MICHAEL:
		0. Normal standing pose, hands in pockets
		1. One hand out, like he's offering/explaining something
		2. Looking to the side, thinking
		3. Pointing at you with one eye visible, addressing you directly
		4. Tipping the brim of his hat, face hidden
		5. Looking away from you

	VANNY:
		0. Pointing up, winking
		1. Looking off to the side, swooning
		2. Confused look, head tilted
		3. Covering her mouth as she giggles
		4. Staring at you blankly
		5. Giving an aside, smugly
		6. Throwing up her arms, looking away from you
		7. Tugging on her ears angrily
		8. It's a surprise.

Hopefully that was comprehensive enough for you... Once more, good luck with your guard modding journey!
Written with love by BubbledDEV <3