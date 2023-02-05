using ElectronicFish.utils;
using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

namespace ElectronicFish
{

	public class WelcomeManager : MonoBehaviour
	{
		[SerializeField] private Text comboText;
		[SerializeField] private Button startGameButton;
		[SerializeField] private Button leaderboardButton;

		private bool _tryAgain = true;

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

#if !UNITY_EDITOR
			PlayGamesPlatform.Activate();
#endif
			Social.localUser.Authenticate(ProcessAuthentication);
		}

		private static void StartGame()
		{
			Debug.Log("nmsl");

			SceneManager.LoadScene("Scenes/MainScene");
		}

		private void ProcessAuthentication(bool status)
		{
			if (status)
			{
				AndroidUtils.ShowAndroidToastMessage("登录成功");
			}
			AndroidUtils.ShowAndroidToastMessage("登录失败");
			if (!_tryAgain)
			{
				leaderboardButton.enabled = false;
				leaderboardButton.GetComponentInChildren<Text>().text = "排行榜 (无法使用)";
			};
			Social.localUser.Authenticate(ProcessAuthentication);
			_tryAgain = false;
		}

		private static void ShowLeaderboard()
		{
			Social.ShowLeaderboardUI();
		}
	}

}
