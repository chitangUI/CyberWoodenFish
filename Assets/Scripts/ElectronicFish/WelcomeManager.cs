using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;
using GooglePlayGames;
using GooglePlayGames.BasicApi;

namespace ElectronicFish
{

	public class WelcomeManager : MonoBehaviour
	{
		[SerializeField] private Text comboText;
		[SerializeField] private Button startGameButton;
		[SerializeField] private Button leaderboardButton;

		private void Awake()
		{
			Application.targetFrameRate = 114514;
		}

		private void Start()
		{
			var combo = PlayerPrefs.GetInt("Merit");
			comboText.text = $"您累计敲了 {combo} 次";

			startGameButton.onClick.AddListener(StartGame);
			leaderboardButton.onClick.AddListener(ShowLeaderboard);


			PlayGamesPlatform.Activate();
			Social.localUser.Authenticate(ProcessAuthentication);
		}

		private static void StartGame()
		{
			Debug.Log("nmsl");

			SceneManager.LoadScene("Scenes/MainScene");
		}

		// ReSharper disable Unity.PerformanceAnalysis
		private static void ProcessAuthentication(bool status)
		{
			if (!status)
			{
				Social.localUser.Authenticate(ProcessAuthentication);
			}
		}

		private static void ShowLeaderboard()
		{
			Social.ShowLeaderboardUI();
		}
	}

}
