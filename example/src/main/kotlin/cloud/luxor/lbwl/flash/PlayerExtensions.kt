package cloud.luxor.lbwl.flash

import com.google.common.io.ByteStreams
import org.bukkit.*
import org.bukkit.entity.Player
import org.bukkit.plugin.Plugin
import org.bukkit.potion.PotionEffect
import org.bukkit.potion.PotionEffectType
import java.util.*


private val gameData = WeakHashMap<Player, GameData>()

val HIDE_PLAYER_ITEM = create(Material.BLAZE_ROD, 0, "§c§lSpieler verstecken §r§7§o<Rechtsklick>")
val SHOW_PLAYER_ITEM = create(Material.STICK, 0, "§c§lSpieler anzeigen §r§7§o<Rechtsklick>")
val RESPAWN_ITEM = create(Material.RED_DYE, 1, "§c§lInstant-Tod(TM) §r§7§o<Rechtsklick>")

fun Player.initGameData(speed: Int, spawn: Location) {
    gameData[player] = GameData(mutableSetOf(), speed, spawn)
}

fun Player.isIngame() = this.gameMode != GameMode.SPECTATOR

fun Player.getCurrentCheckPointLocation() = gameData[this]?.checkpoints?.last()?.location

fun Player.getCurrentCheckPointIndex() = gameData[this]?.checkpoints?.size ?: 0

fun Player.setCurrentCheckpoint(checkpoint: Checkpoint) {
    gameData[this]?.checkpoints?.add(checkpoint)
    println(gameData[this]?.checkpoints?.size)
}

fun Player.respawn() {
    this.applyEffects()

    if (gameData[player]?.checkpoints.isNullOrEmpty()) {
        gameData[player]?.spawn?.let { this.teleport(it) }
    } else {
        this.getCurrentCheckPointLocation()?.let { this.teleport(it) }
    }

    this.playSound(this.location, Sound.ENTITY_ENDERMAN_TELEPORT, 1.0F, 1.0F)

    this.health = 20.0
    this.fireTicks = 0
    this.fallDistance = 0.0F
}

fun Player.sendTweetLink(map: String, time: String) {
    val link = createTwitterLink(
        "§b§l>§r §eTEILE DEINEN REKORD §b§l<",
        "Ich habe einen neuen FLASH-Rekord auf $map erreicht: $time Minuten",
        "FLASHRecords",
        "IntoTheLABS"
    )
    this.sendMessage(link)
}

fun Player.applyEffects() {
    this.activePotionEffects.forEach { this.removePotionEffect(it.type) }
    this.addPotionEffect(PotionEffect(PotionEffectType.REGENERATION, Integer.MAX_VALUE, 3))
    this.addPotionEffect(PotionEffect(PotionEffectType.SPEED, Integer.MAX_VALUE, gameData[player]?.mapSpeed ?: 19 ))
}

fun Player.toggleVisibility(plugin: Plugin) {
    if (this.inventory.getItem(5)?.type == Material.BLAZE_ROD) {
        this.sendMessage("$PREFIX §7Du hast alle Spieler §cversteckt§7.")
        this.playSound(this.location, Sound.BLOCK_NOTE_BLOCK_PLING, 1.0F, 1.0F)
        this.inventory.setItem(5, SHOW_PLAYER_ITEM)
        Bukkit.getOnlinePlayers()
            .filter { it != this }
            .forEach { this.hidePlayer(plugin, it) }
    } else {
        this.inventory.setItem(5, HIDE_PLAYER_ITEM)
        player?.sendMessage("$PREFIX §7Du §asiehst §7nun alle Spieler§7.")
        player?.let { it.playSound(it.location, Sound.BLOCK_NOTE_BLOCK_PLING, 1.0F, 1.0F) }
        Bukkit.getOnlinePlayers()
            .filter { it != player }
            .forEach { player?.showPlayer(plugin, it) }
    }
}

fun Player.giveItems() {
    this.inventory.clear()
    this.inventory.setItem(3, RESPAWN_ITEM)
    this.inventory.setItem(5, HIDE_PLAYER_ITEM)
}

fun Player.connectToLobby(plugin: Plugin) {
    val out = ByteStreams.newDataOutput()
    out.writeUTF("ConnectOther")
    out.writeUTF(this.name)
    out.writeUTF("lobby")
    this.sendPluginMessage(plugin, "BungeeCord", out.toByteArray())
}
