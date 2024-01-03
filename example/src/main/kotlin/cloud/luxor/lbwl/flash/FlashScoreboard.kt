package cloud.luxor.lbwl.flash

import net.kyori.adventure.text.Component
import org.bukkit.Bukkit
import org.bukkit.entity.Player
import org.bukkit.scheduler.BukkitTask
import org.bukkit.scoreboard.Criteria
import org.bukkit.scoreboard.DisplaySlot


/**
 * @author Yannic Rieger
 */
class FlashScoreboard(private val plugin: FlashPlugin) {

    private val handle = Bukkit.getScoreboardManager().newScoreboard
    private var task: BukkitTask? = null

    init {
        val obj = handle.registerNewObjective("scoreboard", Criteria.DUMMY, Component.text("§e>> §6Flash"))
        obj.displaySlot = DisplaySlot.SIDEBAR
        obj.getScore("§e§lServer-IP:").score = 999
        obj.getScore("§fbergwerkLABS.de").score = 998
        obj.getScore("§1§2§3").score = 997
        obj.getScore("§e§lCheckpoints:").score = 996
    }

    fun startDisplay() {
        val obj = handle.getObjective("scoreboard")
        task = Bukkit.getScheduler().runTaskTimer(plugin, Runnable {
            Bukkit.getOnlinePlayers()
                .filter { it.isIngame() }
                .forEach {
                    obj?.getScore(it.name)?.score = it.getCurrentCheckPointIndex()
                }
        }, 0, 10)
    }

    fun show(player: Player) {
        player.scoreboard = this.handle
    }

    fun updateTitle(title: String) {
        handle.getObjective("scoreboard")?.displayName(Component.text("$PREFIX §b$title"))
    }

    fun destroy() {
        this.task?.cancel()
        Bukkit.getOnlinePlayers().forEach { it.scoreboard = Bukkit.getScoreboardManager().newScoreboard }
    }
}
